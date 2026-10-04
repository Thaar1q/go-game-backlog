-- 1. Users
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL       PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL UNIQUE,
    email      VARCHAR(100) NOT NULL UNIQUE,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users (is_active);

-- 2. Games catalog
CREATE TABLE IF NOT EXISTS games (
    id           SERIAL       PRIMARY KEY,
    title        VARCHAR(150) NOT NULL,
    platform     VARCHAR(50)  NOT NULL,
    genre        VARCHAR(50)  NOT NULL,
    release_year INTEGER      NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_games_title_lower ON games (LOWER(title));
CREATE INDEX IF NOT EXISTS idx_games_genre ON games (genre);
CREATE INDEX IF NOT EXISTS idx_games_platform ON games (platform);

-- 3. Personal Backlog
CREATE TABLE IF NOT EXISTS backlogs (
    id           SERIAL        PRIMARY KEY,
    user_id      INTEGER       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id      INTEGER       NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    status       VARCHAR(20)   NOT NULL DEFAULT 'backlog',
    hours_played INT           NOT NULL DEFAULT 0,
    rating       INT           NULL,
    notes        TEXT,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_user_game UNIQUE (user_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_backlogs_user_id ON backlogs (user_id);