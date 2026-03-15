-- Active: 1773538779120@@127.0.0.1@5432@mydb@public

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS events(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ,
    path VARCHAR(255),
    type ENUM('USER_EVENT', 'COMMERCE_EVENT')
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS properties(
    event_id UUID PRIMARY KEY REFERENCES events (id) ON DELETE CASCADE,
    total FLOAT NOT NULL,
    page TEXT NOT NULL
);