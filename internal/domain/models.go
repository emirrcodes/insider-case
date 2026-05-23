package domain

type Team struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Strength int    `json:"strength"`
}

type Match struct {
	ID         int64  `json:"id"`
	Week       int    `json:"week"`
	HomeTeamID int64  `json:"home_team_id"`
	AwayTeamID int64  `json:"away_team_id"`
	HomeTeam   Team   `json:"home_team"`
	AwayTeam   Team   `json:"away_team"`
	HomeScore  *int   `json:"home_score"`
	AwayScore  *int   `json:"away_score"`
	Played     bool   `json:"played"`
	Result     string `json:"result"`
}

type Standing struct {
	Position       int    `json:"position"`
	TeamID         int64  `json:"team_id"`
	Team           string `json:"team"`
	Played         int    `json:"played"`
	Won            int    `json:"won"`
	Drawn          int    `json:"drawn"`
	Lost           int    `json:"lost"`
	GoalsFor       int    `json:"goals_for"`
	GoalsAgainst   int    `json:"goals_against"`
	GoalDifference int    `json:"goal_difference"`
	Points         int    `json:"points"`
}

type Prediction struct {
	TeamID                      int64   `json:"team_id"`
	Team                        string  `json:"team"`
	ChampionshipProbability     float64 `json:"championship_probability"`
	ChampionshipProbabilityText string  `json:"championship_probability_text"`
}

type WeekResult struct {
	Week    int     `json:"week"`
	Matches []Match `json:"matches"`
}

type LeagueSnapshot struct {
	CurrentWeek int          `json:"current_week"`
	TotalWeeks  int          `json:"total_weeks"`
	Standings   []Standing   `json:"standings"`
	Matches     []Match      `json:"matches"`
	Predictions []Prediction `json:"predictions,omitempty"`
}
