package service

import (
	"sort"

	"github.com/ahmetemirarslan/insider-backend-case/internal/domain"
)

func CalculateStandings(teams []domain.Team, matches []domain.Match) []domain.Standing {
	table := make(map[int64]*domain.Standing, len(teams))
	for _, team := range teams {
		table[team.ID] = &domain.Standing{
			TeamID: team.ID,
			Team:   team.Name,
		}
	}

	for _, match := range matches {
		if !match.Played || match.HomeScore == nil || match.AwayScore == nil {
			continue
		}
		ensureStanding(table, match.HomeTeam)
		ensureStanding(table, match.AwayTeam)

		home := table[match.HomeTeamID]
		away := table[match.AwayTeamID]
		homeGoals := *match.HomeScore
		awayGoals := *match.AwayScore

		home.Played++
		away.Played++
		home.GoalsFor += homeGoals
		home.GoalsAgainst += awayGoals
		away.GoalsFor += awayGoals
		away.GoalsAgainst += homeGoals

		switch {
		case homeGoals > awayGoals:
			home.Won++
			away.Lost++
			home.Points += 3
		case homeGoals < awayGoals:
			away.Won++
			home.Lost++
			away.Points += 3
		default:
			home.Drawn++
			away.Drawn++
			home.Points++
			away.Points++
		}
	}

	standings := make([]domain.Standing, 0, len(table))
	for _, standing := range table {
		standing.GoalDifference = standing.GoalsFor - standing.GoalsAgainst
		standings = append(standings, *standing)
	}

	sort.SliceStable(standings, func(i, j int) bool {
		left := standings[i]
		right := standings[j]
		if left.Points != right.Points {
			return left.Points > right.Points
		}
		if left.GoalDifference != right.GoalDifference {
			return left.GoalDifference > right.GoalDifference
		}
		if left.GoalsFor != right.GoalsFor {
			return left.GoalsFor > right.GoalsFor
		}
		return left.Team < right.Team
	})

	for i := range standings {
		standings[i].Position = i + 1
	}

	return standings
}

func ensureStanding(table map[int64]*domain.Standing, team domain.Team) {
	if _, ok := table[team.ID]; ok {
		return
	}
	table[team.ID] = &domain.Standing{
		TeamID: team.ID,
		Team:   team.Name,
	}
}
