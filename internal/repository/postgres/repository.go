package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahmetemirarslan/insider-backend-case/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

type Repository struct {
	*Store
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Store: &Store{pool: pool}}
}

func (r *Repository) ResetLeague(ctx context.Context, teams []domain.Team, matches []domain.Match) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "TRUNCATE TABLE matches, teams RESTART IDENTITY CASCADE"); err != nil {
		return err
	}

	for _, team := range teams {
		if _, err := tx.Exec(ctx,
			"INSERT INTO teams (id, name, strength) VALUES ($1, $2, $3)",
			team.ID,
			team.Name,
			team.Strength,
		); err != nil {
			return err
		}
	}

	for _, match := range matches {
		if _, err := tx.Exec(ctx,
			"INSERT INTO matches (id, week, home_team_id, away_team_id) VALUES ($1, $2, $3, $4)",
			match.ID,
			match.Week,
			match.HomeTeamID,
			match.AwayTeamID,
		); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, "SELECT setval('teams_id_seq', (SELECT MAX(id) FROM teams))"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "SELECT setval('matches_id_seq', (SELECT MAX(id) FROM matches))"); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) ListTeams(ctx context.Context) ([]domain.Team, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, name, strength FROM teams ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]domain.Team, 0)
	for rows.Next() {
		var team domain.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Strength); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *Repository) ListMatches(ctx context.Context) ([]domain.Match, error) {
	return r.listMatches(ctx, "")
}

func (r *Repository) ListMatchesByWeek(ctx context.Context, week int) ([]domain.Match, error) {
	return r.listMatches(ctx, "WHERE m.week = $1", week)
}

func (r *Repository) NextUnplayedWeek(ctx context.Context) (*int, error) {
	var week int
	err := r.pool.QueryRow(ctx, "SELECT week FROM matches WHERE played = false ORDER BY week LIMIT 1").Scan(&week)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &week, nil
}

func (r *Repository) TotalWeeks(ctx context.Context) (int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COALESCE(MAX(week), 0) FROM matches").Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) UpdateMatchResult(ctx context.Context, matchID int64, homeScore int, awayScore int) (domain.Match, error) {
	query := `
		UPDATE matches
		SET home_score = $2,
			away_score = $3,
			played = true,
			updated_at = now()
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, query, matchID, homeScore, awayScore); err != nil {
		return domain.Match{}, err
	}

	rows, err := r.listMatches(ctx, "WHERE m.id = $1", matchID)
	if err != nil {
		return domain.Match{}, err
	}
	if len(rows) == 0 {
		return domain.Match{}, pgx.ErrNoRows
	}
	return rows[0], nil
}

func (r *Repository) listMatches(ctx context.Context, where string, args ...any) ([]domain.Match, error) {
	query := `
		SELECT
			m.id,
			m.week,
			m.home_team_id,
			m.away_team_id,
			m.home_score,
			m.away_score,
			m.played,
			ht.id,
			ht.name,
			ht.strength,
			at.id,
			at.name,
			at.strength
		FROM matches m
		JOIN teams ht ON ht.id = m.home_team_id
		JOIN teams at ON at.id = m.away_team_id
	`
	if where != "" {
		query = fmt.Sprintf("%s %s", query, where)
	}
	query = fmt.Sprintf("%s ORDER BY m.week, m.id", query)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]domain.Match, 0)
	for rows.Next() {
		match, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, rows.Err()
}

func scanMatch(rows pgx.Rows) (domain.Match, error) {
	var match domain.Match
	var homeScore pgtype.Int4
	var awayScore pgtype.Int4

	err := rows.Scan(
		&match.ID,
		&match.Week,
		&match.HomeTeamID,
		&match.AwayTeamID,
		&homeScore,
		&awayScore,
		&match.Played,
		&match.HomeTeam.ID,
		&match.HomeTeam.Name,
		&match.HomeTeam.Strength,
		&match.AwayTeam.ID,
		&match.AwayTeam.Name,
		&match.AwayTeam.Strength,
	)
	if err != nil {
		return domain.Match{}, err
	}

	if homeScore.Valid {
		score := int(homeScore.Int32)
		match.HomeScore = &score
	}
	if awayScore.Valid {
		score := int(awayScore.Int32)
		match.AwayScore = &score
	}
	if match.Played && match.HomeScore != nil && match.AwayScore != nil {
		match.Result = fmt.Sprintf("%s %d - %d %s", match.HomeTeam.Name, *match.HomeScore, *match.AwayScore, match.AwayTeam.Name)
	}

	return match, nil
}
