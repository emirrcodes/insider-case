CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    strength INTEGER NOT NULL CHECK (strength BETWEEN 1 AND 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS matches (
    id BIGSERIAL PRIMARY KEY,
    week INTEGER NOT NULL CHECK (week > 0),
    home_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    away_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    home_score INTEGER CHECK (home_score >= 0),
    away_score INTEGER CHECK (away_score >= 0),
    played BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT matches_distinct_teams CHECK (home_team_id <> away_team_id),
    CONSTRAINT matches_score_state CHECK (
        (played = false AND home_score IS NULL AND away_score IS NULL)
        OR
        (played = true AND home_score IS NOT NULL AND away_score IS NOT NULL)
    ),
    CONSTRAINT matches_unique_fixture UNIQUE (week, home_team_id, away_team_id)
);

CREATE INDEX IF NOT EXISTS idx_matches_week ON matches(week);
CREATE INDEX IF NOT EXISTS idx_matches_played_week ON matches(played, week);
CREATE INDEX IF NOT EXISTS idx_matches_home_team ON matches(home_team_id);
CREATE INDEX IF NOT EXISTS idx_matches_away_team ON matches(away_team_id);
