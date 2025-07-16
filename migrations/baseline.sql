CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    uid UUID NOT NULL, 
    name text NOT NULL,
    cost INTEGER NOT NULL,
    created_at DATE CHECK (EXTRACT(DAY FROM created_at) = 1) NOT NULL,
    expires TIMESTAMPTZ
);