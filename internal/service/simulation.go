package service

import (
	"math"
	"math/rand"

	"github.com/ahmetemirarslan/insider-backend-case/internal/domain"
)

type MatchSimulator interface {
	Simulate(match domain.Match) (int, int)
}

type DeterministicSimulator struct {
	Seed int64
}

func (s DeterministicSimulator) Simulate(match domain.Match) (int, int) {
	seed := s.Seed + match.ID*10_007 + int64(match.Week)*1_009 + match.HomeTeamID*97 + match.AwayTeamID*53
	rng := rand.New(rand.NewSource(seed))
	return simulateScore(match.HomeTeam.Strength, match.AwayTeam.Strength, rng)
}

func simulateScore(homeStrength int, awayStrength int, rng *rand.Rand) (int, int) {
	strengthDelta := float64(homeStrength-awayStrength) / 35.0
	homeLambda := clamp(1.45+0.20+strengthDelta, 0.25, 4.25)
	awayLambda := clamp(1.20-strengthDelta, 0.20, 4.00)

	return capGoals(poisson(homeLambda, rng)), capGoals(poisson(awayLambda, rng))
}

func poisson(lambda float64, rng *rand.Rand) int {
	limit := math.Exp(-lambda)
	k := 0
	product := 1.0
	for product > limit {
		k++
		product *= rng.Float64()
	}
	return k - 1
}

func clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func capGoals(value int) int {
	if value < 0 {
		return 0
	}
	if value > 7 {
		return 7
	}
	return value
}
