ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(255) NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_transactions_idempotency_key
    ON transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
