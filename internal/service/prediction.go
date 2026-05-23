package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/ahmetemirarslan/insider-backend-case/internal/domain"
)

type Predictor interface {
	Predict(ctx context.Context, teams []domain.Team, matches []domain.Match) []domain.Prediction
}

type MonteCarloPredictor struct {
	Iterations int
	Seed       int64
}

func (p MonteCarloPredictor) Predict(ctx context.Context, teams []domain.Team, matches []domain.Match) []domain.Prediction {
	iterations := p.Iterations
	if iterations <= 0 {
		iterations = 10000
	}

	wins := make(map[int64]int, len(teams))
	names := make(map[int64]string, len(teams))
	for _, team := range teams {
		names[team.ID] = team.Name
	}

	for i := 0; i < iterations; i++ {
		select {
		case <-ctx.Done():
			return formatPredictions(wins, names, i)
		default:
		}

		projected := make([]domain.Match, len(matches))
		copy(projected, matches)

		for index := range projected {
			if projected[index].Played {
				continue
			}
			seed := p.Seed + int64(i+1)*100_003 + projected[index].ID*1_009 + int64(projected[index].Week)*251
			rng := rand.New(rand.NewSource(seed))
			homeScore, awayScore := simulateScore(projected[index].HomeTeam.Strength, projected[index].AwayTeam.Strength, rng)
			projected[index].Played = true
			projected[index].HomeScore = intPointer(homeScore)
			projected[index].AwayScore = intPointer(awayScore)
		}

		standings := CalculateStandings(teams, projected)
		if len(standings) > 0 {
			wins[standings[0].TeamID]++
		}
	}

	return formatPredictions(wins, names, iterations)
}

func formatPredictions(wins map[int64]int, names map[int64]string, total int) []domain.Prediction {
	predictions := make([]domain.Prediction, 0, len(names))
	for teamID, name := range names {
		probability := 0.0
		if total > 0 {
			probability = math.Round((float64(wins[teamID])/float64(total))*10000) / 100
		}
		predictions = append(predictions, domain.Prediction{
			TeamID:                      teamID,
			Team:                        name,
			ChampionshipProbability:     probability,
			ChampionshipProbabilityText: fmt.Sprintf("%%%0.2f", probability),
		})
	}

	sort.SliceStable(predictions, func(i, j int) bool {
		if predictions[i].ChampionshipProbability != predictions[j].ChampionshipProbability {
			return predictions[i].ChampionshipProbability > predictions[j].ChampionshipProbability
		}
		return predictions[i].Team < predictions[j].Team
	})

	return predictions
}

func intPointer(value int) *int {
	return &value
}
