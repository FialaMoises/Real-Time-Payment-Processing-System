-- Webhook Subscriptions
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255) NOT NULL, -- HMAC secret for signature
    events TEXT[] NOT NULL, -- Array of event types
    is_active BOOLEAN NOT NULL DEFAULT true,
    retry_count INTEGER NOT NULL DEFAULT 3,
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Webhook Events/Logs
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL REFERENCES webhook_subscriptions(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    event_id UUID NOT NULL, -- ID of the transaction/alert/etc
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'SUCCESS', 'FAILED', 'RETRYING'
    attempt_count INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP,
    response_status_code INTEGER,
    response_body TEXT,
    error_message TEXT,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Webhook Event Types (for reference)
CREATE TABLE IF NOT EXISTS webhook_event_types (
    event_type VARCHAR(100) PRIMARY KEY,
    description TEXT NOT NULL,
    payload_schema JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices
CREATE INDEX idx_webhook_subscriptions_user ON webhook_subscriptions(user_id);
CREATE INDEX idx_webhook_subscriptions_active ON webhook_subscriptions(is_active);
CREATE INDEX idx_webhook_deliveries_subscription ON webhook_deliveries(subscription_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_event ON webhook_deliveries(event_type, event_id);
CREATE INDEX idx_webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE status = 'RETRYING';

-- Triggers
CREATE TRIGGER update_webhook_subscriptions_updated_at
    BEFORE UPDATE ON webhook_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhook_deliveries_updated_at
    BEFORE UPDATE ON webhook_deliveries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default event types
INSERT INTO webhook_event_types (event_type, description, payload_schema) VALUES
('transaction.created', 'Triggered when a new transaction is created', '{"transaction_id": "uuid", "type": "string", "amount": "float", "status": "string"}'::jsonb),
('transaction.completed', 'Triggered when a transaction is completed', '{"transaction_id": "uuid", "type": "string", "amount": "float", "status": "string"}'::jsonb),
('transaction.failed', 'Triggered when a transaction fails', '{"transaction_id": "uuid", "type": "string", "amount": "float", "error": "string"}'::jsonb),
('fraud.detected', 'Triggered when fraud is detected', '{"transaction_id": "uuid", "alert_id": "uuid", "risk_score": "int", "reasons": "array"}'::jsonb),
('fraud.confirmed', 'Triggered when fraud alert is confirmed', '{"alert_id": "uuid", "transaction_id": "uuid"}'::jsonb),
('account.created', 'Triggered when account is created', '{"account_id": "uuid", "user_id": "uuid"}'::jsonb),
('account.balance_low', 'Triggered when account balance is low', '{"account_id": "uuid", "balance": "float", "threshold": "float"}'::jsonb);
