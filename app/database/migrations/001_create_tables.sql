CREATE TABLE IF NOT EXISTS leagues (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    team_count INTEGER NOT NULL,
    max_weeks INTEGER NOT NULL,
    curr_week INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    league_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    strength INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS team_stats (
    team_id INTEGER PRIMARY KEY,
    points INTEGER DEFAULT 0,
    played INTEGER DEFAULT 0,
    won INTEGER DEFAULT 0,
    lost INTEGER DEFAULT 0,
    draw INTEGER DEFAULT 0,
    goals_for INTEGER DEFAULT 0,
    goals_against INTEGER DEFAULT 0,
    goal_diff INTEGER DEFAULT 0,
    estimation REAL DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    league_id INTEGER NOT NULL,
    week INTEGER NOT NULL,
    played BOOLEAN DEFAULT false,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_score INTEGER DEFAULT 0,
    away_score INTEGER DEFAULT 0,
    result INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS weekly_logs (
    id SERIAL PRIMARY KEY,
    league_id INTEGER NOT NULL,
    week INTEGER NOT NULL,
    team_stats_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_teams_league'
    ) THEN
        ALTER TABLE teams
        ADD CONSTRAINT fk_teams_league
        FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_team_stats_team'
    ) THEN
        ALTER TABLE team_stats
        ADD CONSTRAINT fk_team_stats_team
        FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_matches_league'
    ) THEN
        ALTER TABLE matches
        ADD CONSTRAINT fk_matches_league
        FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_matches_home_team'
    ) THEN
        ALTER TABLE matches
        ADD CONSTRAINT fk_matches_home_team
        FOREIGN KEY (home_team_id) REFERENCES teams(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_matches_away_team'
    ) THEN
        ALTER TABLE matches
        ADD CONSTRAINT fk_matches_away_team
        FOREIGN KEY (away_team_id) REFERENCES teams(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_matches_result'
    ) THEN
        ALTER TABLE matches
        ADD CONSTRAINT fk_matches_result
        FOREIGN KEY (result) REFERENCES teams(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_weekly_logs_league'
    ) THEN
        ALTER TABLE weekly_logs
        ADD CONSTRAINT fk_weekly_logs_league
        FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE;
    END IF;
END $$;
