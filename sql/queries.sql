-- Reset the league state.
TRUNCATE TABLE matches, teams RESTART IDENTITY CASCADE;

-- Seed teams.
INSERT INTO teams (id, name, strength) VALUES
    (1, 'Chelsea', 90),
    (2, 'Arsenal', 82),
    (3, 'Manchester City', 88),
    (4, 'Liverpool', 75);

-- Seed the double round-robin fixture.
INSERT INTO matches (id, week, home_team_id, away_team_id) VALUES
    (1, 1, 1, 4),
    (2, 1, 2, 3),
    (3, 2, 1, 3),
    (4, 2, 4, 2),
    (5, 3, 1, 2),
    (6, 3, 3, 4),
    (7, 4, 4, 1),
    (8, 4, 3, 2),
    (9, 5, 3, 1),
    (10, 5, 2, 4),
    (11, 6, 2, 1),
    (12, 6, 4, 3);

-- List teams.
SELECT id, name, strength
FROM teams
ORDER BY id;

-- List matches with team details.
SELECT
    m.id,
    m.week,
    m.home_team_id,
    m.away_team_id,
    m.home_score,
    m.away_score,
    m.played,
    ht.id AS home_id,
    ht.name AS home_name,
    ht.strength AS home_strength,
    at.id AS away_id,
    at.name AS away_name,
    at.strength AS away_strength
FROM matches m
JOIN teams ht ON ht.id = m.home_team_id
JOIN teams at ON at.id = m.away_team_id
ORDER BY m.week, m.id;

-- List a single week.
SELECT *
FROM matches
WHERE week = 4
ORDER BY id;

-- Find the next week that has unplayed matches.
SELECT week
FROM matches
WHERE played = false
ORDER BY week
LIMIT 1;

-- Save or edit a match result. Standings are recalculated from matches by the service.
UPDATE matches
SET home_score = 2,
    away_score = 1,
    played = true,
    updated_at = now()
WHERE id = 1;
