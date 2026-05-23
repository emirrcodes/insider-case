package repository

import (
	"context"

	"github.com/emirrcodes/insider-case/internal/domain"
)

type LeagueRepository interface {
	ResetLeague(ctx context.Context, teams []domain.Team, matches []domain.Match) error
	ListTeams(ctx context.Context) ([]domain.Team, error)
	ListMatches(ctx context.Context) ([]domain.Match, error)
	ListMatchesByWeek(ctx context.Context, week int) ([]domain.Match, error)
	NextUnplayedWeek(ctx context.Context) (*int, error)
	TotalWeeks(ctx context.Context) (int, error)
	UpdateMatchResult(ctx context.Context, matchID int64, homeScore int, awayScore int) (domain.Match, error)
}
