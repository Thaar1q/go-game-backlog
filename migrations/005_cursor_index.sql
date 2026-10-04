CREATE INDEX IF NOT EXISTS idx_games_created_at_id_desc
    ON games (created_at DESC, id DESC);