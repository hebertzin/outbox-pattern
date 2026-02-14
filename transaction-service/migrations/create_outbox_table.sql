CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    type VARCHAR(200) NOT NULL,
    payload TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_status
    ON outbox (status);

CREATE INDEX IF NOT EXISTS idx_outbox_created_at
    ON outbox (created_at);
