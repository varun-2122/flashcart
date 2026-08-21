-- 000006: payments table
-- Stores one payment attempt per order. IdempotencyKey prevents duplicate charges on client retries.

CREATE TABLE IF NOT EXISTS payments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id         UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount           NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    status           VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    provider         VARCHAR(100) NOT NULL DEFAULT 'simulated',
    idempotency_key  VARCHAR(255) UNIQUE NOT NULL,
    transaction_id   VARCHAR(255),
    failure_reason   TEXT,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payments_order     ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_user      ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_idem_key  ON payments(idempotency_key);
