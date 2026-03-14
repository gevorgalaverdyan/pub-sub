CREATE TABLE IF NOT EXISTS events(
    id UUID PRIMARY KEY gen_random_uuid(),
    event VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ,
    path VARCHAR(255),
    created_at  TIMESTAMPTZ DEFAULT now()
)

CREATE TABLE IF NOT EXISTS properties(
    event_id UUID PRIMARY KEY REFERENCES events(id) ON DELETE CASCADE,
    total FLOAT NOT NULL,
    page TEXT NOT NULL,
)