package service

import (
	"context"
	"errors"

	"github.com/emirrcodes/insider-case/internal/domain"
	"github.com/emirrcodes/insider-case/internal/repository"
	"github.com/jackc/pgx/v5"
)

var (
	ErrLeagueFinished = errors.New("all league matches have already been played")
	ErrInvalidScore   = errors.New("scores must be zero or positive")
	ErrMatchNotFound  = errors.New("match not found")
)

type LeagueService struct {
	repository repository.LeagueRepository
	simulator  MatchSimulator
	predictor  Predictor
}

func NewLeagueService(repository repository.LeagueRepository, simulator MatchSimulator, predictor Predictor) *LeagueService {
	return &LeagueService{
		repository: repository,
		simulator:  simulator,
		predictor:  predictor,
	}
}

func (s *LeagueService) Reset(ctx context.Context) (domain.LeagueSnapshot, error) {
	if err := s.repository.ResetLeague(ctx, defaultTeams(), defaultSchedule()); err != nil {
		return domain.LeagueSnapshot{}, err
	}
	return s.Snapshot(ctx)
}

func (s *LeagueService) Snapshot(ctx context.Context) (domain.LeagueSnapshot, error) {
	teams, matches, totalWeeks, err := s.loadLeague(ctx)
	if err != nil {
		return domain.LeagueSnapshot{}, err
	}

	standings := CalculateStandings(teams, matches)
	currentWeek := completedWeeks(matches)
	snapshot := domain.LeagueSnapshot{
		CurrentWeek: currentWeek,
		TotalWeeks:  totalWeeks,
		Standings:   standings,
		Matches:     matches,
	}

	if currentWeek >= 4 {
		snapshot.Predictions = s.predictor.Predict(ctx, teams, matches)
	}

	return snapshot, nil
}

func (s *LeagueService) Standings(ctx context.Context) ([]domain.Standing, error) {
	teams, matches, _, err := s.loadLeague(ctx)
	if err != nil {
		return nil, err
	}
	return CalculateStandings(teams, matches), nil
}

func (s *LeagueService) Matches(ctx context.Context) ([]domain.Match, error) {
	return s.repository.ListMatches(ctx)
}

func (s *LeagueService) MatchesByWeek(ctx context.Context, week int) ([]domain.Match, error) {
	return s.repository.ListMatchesByWeek(ctx, week)
}

func (s *LeagueService) Predictions(ctx context.Context) ([]domain.Prediction, error) {
	teams, matches, _, err := s.loadLeague(ctx)
	if err != nil {
		return nil, err
	}
	if completedWeeks(matches) < 4 {
		return []domain.Prediction{}, nil
	}
	return s.predictor.Predict(ctx, teams, matches), nil
}

func (s *LeagueService) PlayNextWeek(ctx context.Context) (domain.WeekResult, error) {
	week, err := s.repository.NextUnplayedWeek(ctx)
	if err != nil {
		return domain.WeekResult{}, err
	}
	if week == nil {
		return domain.WeekResult{}, ErrLeagueFinished
	}

	matches, err := s.repository.ListMatchesByWeek(ctx, *week)
	if err != nil {
		return domain.WeekResult{}, err
	}

	played := make([]domain.Match, 0, len(matches))
	for _, match := range matches {
		if !match.Played {
			homeScore, awayScore := s.simulator.Simulate(match)
			match, err = s.repository.UpdateMatchResult(ctx, match.ID, homeScore, awayScore)
			if err != nil {
				return domain.WeekResult{}, err
			}
		}
		played = append(played, match)
	}

	return domain.WeekResult{Week: *week, Matches: played}, nil
}

func (s *LeagueService) PlayAll(ctx context.Context) ([]domain.WeekResult, error) {
	results := make([]domain.WeekResult, 0)
	for {
		result, err := s.PlayNextWeek(ctx)
		if errors.Is(err, ErrLeagueFinished) {
			return results, nil
		}
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
}

func (s *LeagueService) EditMatchResult(ctx context.Context, matchID int64, homeScore int, awayScore int) (domain.LeagueSnapshot, error) {
	if homeScore < 0 || awayScore < 0 {
		return domain.LeagueSnapshot{}, ErrInvalidScore
	}
	if _, err := s.repository.UpdateMatchResult(ctx, matchID, homeScore, awayScore); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LeagueSnapshot{}, ErrMatchNotFound
		}
		return domain.LeagueSnapshot{}, err
	}
	return s.Snapshot(ctx)
}

func (s *LeagueService) loadLeague(ctx context.Context) ([]domain.Team, []domain.Match, int, error) {
	teams, err := s.repository.ListTeams(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	matches, err := s.repository.ListMatches(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	totalWeeks, err := s.repository.TotalWeeks(ctx)
	if err != nil {
		return nil, nil, 0, err
	}
	return teams, matches, totalWeeks, nil
}

func defaultTeams() []domain.Team {
	return []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 90},
		{ID: 2, Name: "Arsenal", Strength: 82},
		{ID: 3, Name: "Manchester City", Strength: 88},
		{ID: 4, Name: "Liverpool", Strength: 75},
	}
}

func defaultSchedule() []domain.Match {
	return []domain.Match{
		{ID: 1, Week: 1, HomeTeamID: 1, AwayTeamID: 4},
		{ID: 2, Week: 1, HomeTeamID: 2, AwayTeamID: 3},
		{ID: 3, Week: 2, HomeTeamID: 1, AwayTeamID: 3},
		{ID: 4, Week: 2, HomeTeamID: 4, AwayTeamID: 2},
		{ID: 5, Week: 3, HomeTeamID: 1, AwayTeamID: 2},
		{ID: 6, Week: 3, HomeTeamID: 3, AwayTeamID: 4},
		{ID: 7, Week: 4, HomeTeamID: 4, AwayTeamID: 1},
		{ID: 8, Week: 4, HomeTeamID: 3, AwayTeamID: 2},
		{ID: 9, Week: 5, HomeTeamID: 3, AwayTeamID: 1},
		{ID: 10, Week: 5, HomeTeamID: 2, AwayTeamID: 4},
		{ID: 11, Week: 6, HomeTeamID: 2, AwayTeamID: 1},
		{ID: 12, Week: 6, HomeTeamID: 4, AwayTeamID: 3},
	}
}

func completedWeeks(matches []domain.Match) int {
	weeks := map[int][]domain.Match{}
	maxWeek := 0
	for _, match := range matches {
		weeks[match.Week] = append(weeks[match.Week], match)
		if match.Week > maxWeek {
			maxWeek = match.Week
		}
	}

	completed := 0
	for week := 1; week <= maxWeek; week++ {
		weekMatches := weeks[week]
		if len(weekMatches) == 0 {
			break
		}
		for _, match := range weekMatches {
			if !match.Played {
				return completed
			}
		}
		completed++
	}
	return completed
}
