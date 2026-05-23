package service

import (
	"testing"

	"github.com/emirrcodes/insider-case/internal/domain"
)

func TestCalculateStandingsSortsByPremierLeagueRules(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea"},
		{ID: 2, Name: "Arsenal"},
		{ID: 3, Name: "Manchester City"},
	}
	chelseaGoals := 2
	arsenalGoals := 0
	cityGoals := 3
	arsenalSecondGoals := 1
	matches := []domain.Match{
		{HomeTeamID: 1, AwayTeamID: 2, HomeTeam: teams[0], AwayTeam: teams[1], HomeScore: &chelseaGoals, AwayScore: &arsenalGoals, Played: true},
		{HomeTeamID: 3, AwayTeamID: 2, HomeTeam: teams[2], AwayTeam: teams[1], HomeScore: &cityGoals, AwayScore: &arsenalSecondGoals, Played: true},
	}

	standings := CalculateStandings(teams, matches)

	if standings[0].Team != "Manchester City" {
		t.Fatalf("expected Manchester City first by goals scored, got %s", standings[0].Team)
	}
	if standings[2].Points != 0 || standings[2].Lost != 2 {
		t.Fatalf("expected Arsenal to have two losses and zero points, got %+v", standings[2])
	}
}
