DROP TRIGGER IF EXISTS update_fraud_alerts_updated_at ON fraud_alerts;
DROP TRIGGER IF EXISTS update_fraud_rules_updated_at ON fraud_rules;

DROP INDEX IF EXISTS idx_fraud_rules_active;
DROP INDEX IF EXISTS idx_fraud_scores_created;
DROP INDEX IF EXISTS idx_fraud_scores_account;
DROP INDEX IF EXISTS idx_fraud_alerts_created;
DROP INDEX IF EXISTS idx_fraud_alerts_status;
DROP INDEX IF EXISTS idx_fraud_alerts_transaction;

DROP TABLE IF EXISTS fraud_scores;
DROP TABLE IF EXISTS fraud_alerts;
DROP TABLE IF EXISTS fraud_rules;
