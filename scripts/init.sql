-- Initial database setup
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    phone VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_cpf ON users(cpf);

-- Accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    account_number VARCHAR(20) UNIQUE NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT positive_balance CHECK (balance >= 0)
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_account_number ON accounts(account_number);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    from_account_id UUID REFERENCES accounts(id),
    to_account_id UUID REFERENCES accounts(id),
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'BRL',
    status VARCHAR(20) NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,

    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT valid_transfer CHECK (
        (type = 'TRANSFER' AND from_account_id IS NOT NULL AND to_account_id IS NOT NULL)
        OR (type = 'DEPOSIT' AND to_account_id IS NOT NULL)
        OR (type = 'WITHDRAWAL' AND from_account_id IS NOT NULL)
    )
);

CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_idempotency ON transactions(idempotency_key);

-- Ledger table (immutable, append-only)
CREATE TABLE IF NOT EXISTS ledger (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    amount DECIMAL(15,2) NOT NULL,
    balance_after DECIMAL(15,2) NOT NULL,
    operation_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT no_zero_amount CHECK (amount != 0)
);

CREATE INDEX idx_ledger_account_id ON ledger(account_id);
CREATE INDEX idx_ledger_transaction_id ON ledger(transaction_id);
CREATE INDEX idx_ledger_created_at ON ledger(created_at DESC);

-- Prevent UPDATE and DELETE on ledger (immutable)
CREATE OR REPLACE RULE ledger_no_update AS
    ON UPDATE TO ledger
    DO INSTEAD NOTHING;

CREATE OR REPLACE RULE ledger_no_delete AS
    ON DELETE TO ledger
    DO INSTEAD NOTHING;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
