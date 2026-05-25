CREATE TABLE teams (
    id       SERIAL PRIMARY KEY,
    name     TEXT UNIQUE NOT NULL,
    strength INT NOT NULL CHECK (strength BETWEEN 1 AND 100)
);

CREATE TABLE matches (
    id           SERIAL PRIMARY KEY,
    week         INT NOT NULL,
    home_team_id INT NOT NULL REFERENCES teams(id),
    away_team_id INT NOT NULL REFERENCES teams(id),
    home_goals   INT,
    away_goals   INT,
    played       BOOLEAN NOT NULL DEFAULT FALSE,
    CHECK (home_team_id <> away_team_id)
);

CREATE INDEX idx_matches_week ON matches(week);
CREATE INDEX idx_matches_played ON matches(played);
