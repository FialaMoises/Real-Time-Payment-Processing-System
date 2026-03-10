-- Fraud Detection Rules
CREATE TABLE IF NOT EXISTS fraud_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL, -- 'HIGH_AMOUNT', 'VELOCITY', 'TIME_PATTERN', 'LOCATION', 'CUSTOM'
    condition JSONB NOT NULL, -- Condições da regra em JSON
    risk_score INTEGER NOT NULL DEFAULT 0, -- Score de risco (0-100)
    action VARCHAR(50) NOT NULL DEFAULT 'FLAG', -- 'FLAG', 'BLOCK', 'REVIEW'
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Fraud Alerts
CREATE TABLE IF NOT EXISTS fraud_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    rule_id UUID REFERENCES fraud_rules(id),
    risk_score INTEGER NOT NULL, -- Score total de risco
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'CONFIRMED', 'FALSE_POSITIVE', 'RESOLVED'
    reason TEXT, -- Motivo da detecção
    details JSONB, -- Detalhes adicionais
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Fraud Scores History (para ML)
CREATE TABLE IF NOT EXISTS fraud_scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    amount_score INTEGER DEFAULT 0,
    velocity_score INTEGER DEFAULT 0,
    time_score INTEGER DEFAULT 0,
    pattern_score INTEGER DEFAULT 0,
    ml_score INTEGER DEFAULT 0, -- Score do modelo de ML
    total_score INTEGER NOT NULL,
    features JSONB, -- Features usadas para cálculo
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para performance
CREATE INDEX idx_fraud_alerts_transaction ON fraud_alerts(transaction_id);
CREATE INDEX idx_fraud_alerts_status ON fraud_alerts(status);
CREATE INDEX idx_fraud_alerts_created ON fraud_alerts(created_at);
CREATE INDEX idx_fraud_scores_account ON fraud_scores(account_id);
CREATE INDEX idx_fraud_scores_created ON fraud_scores(created_at);
CREATE INDEX idx_fraud_rules_active ON fraud_rules(is_active);

-- Trigger para atualizar updated_at
CREATE TRIGGER update_fraud_rules_updated_at
    BEFORE UPDATE ON fraud_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_fraud_alerts_updated_at
    BEFORE UPDATE ON fraud_alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Inserir regras padrão de detecção de fraude
INSERT INTO fraud_rules (name, description, rule_type, condition, risk_score, action) VALUES
('High Amount Transaction', 'Transações acima de R$ 10.000', 'HIGH_AMOUNT', '{"threshold": 10000}', 40, 'FLAG'),
('Very High Amount Transaction', 'Transações acima de R$ 50.000', 'HIGH_AMOUNT', '{"threshold": 50000}', 70, 'REVIEW'),
('Extreme Amount Transaction', 'Transações acima de R$ 100.000', 'HIGH_AMOUNT', '{"threshold": 100000}', 90, 'BLOCK'),
('High Velocity - 5 in 5min', 'Mais de 5 transações em 5 minutos', 'VELOCITY', '{"count": 5, "window_minutes": 5}', 60, 'FLAG'),
('High Velocity - 10 in 1hour', 'Mais de 10 transações em 1 hora', 'VELOCITY', '{"count": 10, "window_minutes": 60}', 50, 'FLAG'),
('Night Time Transaction', 'Transações entre 2h e 5h da manhã', 'TIME_PATTERN', '{"start_hour": 2, "end_hour": 5}', 30, 'FLAG'),
('Rapid Succession', 'Transações consecutivas em menos de 10 segundos', 'VELOCITY', '{"count": 2, "window_seconds": 10}', 50, 'FLAG');
