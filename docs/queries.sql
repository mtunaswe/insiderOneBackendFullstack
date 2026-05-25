-- ============================================================
-- Insider League Simulator — SQL Queries Reference
-- ============================================================
-- All queries used by the application's repository layer.
-- Schema is defined in migrations/0001_init.up.sql
-- Seed data is in migrations/0002_seed_teams.up.sql
-- ============================================================

-- ============================================================
-- TEAM QUERIES
-- ============================================================

-- List all teams
SELECT id, name, strength FROM teams ORDER BY id;

-- Get team by ID
SELECT id, name, strength FROM teams WHERE id = $1;

-- ============================================================
-- MATCH QUERIES
-- ============================================================

-- List all matches (with optional filters)
SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played
FROM matches
ORDER BY week, id;

-- List matches filtered by week
SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played
FROM matches
WHERE week = $1
ORDER BY week, id;

-- List matches filtered by played status
SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played
FROM matches
WHERE played = $1
ORDER BY week, id;

-- List matches filtered by week AND played status
SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played
FROM matches
WHERE week = $1 AND played = $2
ORDER BY week, id;

-- Get match by ID
SELECT id, week, home_team_id, away_team_id, home_goals, away_goals, played
FROM matches
WHERE id = $1;

-- Create matches in batch (fixture generation)
INSERT INTO matches (week, home_team_id, away_team_id, home_goals, away_goals, played)
VALUES
    ($1, $2, $3, $4, $5, $6),
    ($7, $8, $9, $10, $11, $12);
    -- ... repeated for each match in the batch

-- Update match result (simulate or edit)
UPDATE matches
SET home_goals = $1, away_goals = $2, played = true
WHERE id = $3;

-- Truncate matches (league reset)
TRUNCATE matches RESTART IDENTITY;
