BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount              BIGINT NOT NULL CHECK (amount >= 0),
    description         TEXT NOT NULL,
    from_user_id        UUID NOT NULL,
    to_user_id          UUID NOT NULL,
    transaction_status  VARCHAR(50) NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at        TIMESTAMP NULL,

    CONSTRAINT chk_transactions_users_different
    CHECK (from_user_id <> to_user_id)
    );

CREATE INDEX IF NOT EXISTS idx_transactions_from_user_id
    ON transactions (from_user_id);

CREATE INDEX IF NOT EXISTS idx_transactions_to_user_id
    ON transactions (to_user_id);

CREATE INDEX IF NOT EXISTS idx_transactions_created_at
    ON transactions (created_at);

CREATE INDEX IF NOT EXISTS idx_transactions_status
    ON transactions (transaction_status);


COMMIT;
