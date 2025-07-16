CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    uid UUID NOT NULL, 
    name text NOT NULL,
    cost INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires TIMESTAMPTZ
);