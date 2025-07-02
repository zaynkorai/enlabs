CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    balance NUMERIC(20, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    transaction_id VARCHAR(255) UNIQUE NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    state VARCHAR(10) NOT NULL,
    amount NUMERIC(20, 2) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
            REFERENCES users(id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transactions_transaction_id ON transactions (transaction_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);

-- Insert predefined users (id 1, 2, 3) with initial balance 0.00 if they don't exist
INSERT INTO users (id, balance) VALUES (1, 0.00) ON CONFLICT (id) DO NOTHING;
INSERT INTO users (id, balance) VALUES (2, 0.00) ON CONFLICT (id) DO NOTHING;
INSERT INTO users (id, balance) VALUES (3, 0.00) ON CONFLICT (id) DO NOTHING;