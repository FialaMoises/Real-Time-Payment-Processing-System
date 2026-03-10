package fraud

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	// Rules
	GetActiveRules(ctx context.Context) ([]FraudRule, error)
	GetRuleByID(ctx context.Context, id uuid.UUID) (*FraudRule, error)

	// Alerts
	CreateAlert(ctx context.Context, alert *FraudAlert) error
	GetAlertByTransactionID(ctx context.Context, txID uuid.UUID) (*FraudAlert, error)
	UpdateAlertStatus(ctx context.Context, id uuid.UUID, status AlertStatus, reviewedBy uuid.UUID) error
	ListAlerts(ctx context.Context, status AlertStatus, limit, offset int) ([]FraudAlert, error)

	// Scores
	CreateScore(ctx context.Context, score *FraudScore) error
	GetScoresByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]FraudScore, error)

	// Transaction Analysis
	GetRecentTransactions(ctx context.Context, accountID uuid.UUID, since time.Time) ([]RecentTransaction, error)
	GetTransactionCount(ctx context.Context, accountID uuid.UUID, since time.Time) (int, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetActiveRules(ctx context.Context) ([]FraudRule, error) {
	query := `
		SELECT id, name, description, rule_type, condition, risk_score, action, is_active, created_at, updated_at
		FROM fraud_rules
		WHERE is_active = true
		ORDER BY risk_score DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []FraudRule
	for rows.Next() {
		var rule FraudRule
		var conditionJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
			&conditionJSON, &rule.RiskScore, &rule.Action, &rule.IsActive,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(conditionJSON, &rule.Condition); err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func (r *repository) GetRuleByID(ctx context.Context, id uuid.UUID) (*FraudRule, error) {
	query := `
		SELECT id, name, description, rule_type, condition, risk_score, action, is_active, created_at, updated_at
		FROM fraud_rules
		WHERE id = $1
	`

	var rule FraudRule
	var conditionJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
		&conditionJSON, &rule.RiskScore, &rule.Action, &rule.IsActive,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(conditionJSON, &rule.Condition); err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *repository) CreateAlert(ctx context.Context, alert *FraudAlert) error {
	detailsJSON, err := json.Marshal(alert.Details)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO fraud_alerts (id, transaction_id, rule_id, risk_score, status, reason, details, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	alert.ID = uuid.New()
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		alert.ID, alert.TransactionID, alert.RuleID, alert.RiskScore,
		alert.Status, alert.Reason, detailsJSON, alert.CreatedAt, alert.UpdatedAt,
	)

	return err
}

func (r *repository) GetAlertByTransactionID(ctx context.Context, txID uuid.UUID) (*FraudAlert, error) {
	query := `
		SELECT id, transaction_id, rule_id, risk_score, status, reason, details, reviewed_by, reviewed_at, created_at, updated_at
		FROM fraud_alerts
		WHERE transaction_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var alert FraudAlert
	var detailsJSON []byte

	err := r.db.QueryRowContext(ctx, query, txID).Scan(
		&alert.ID, &alert.TransactionID, &alert.RuleID, &alert.RiskScore,
		&alert.Status, &alert.Reason, &detailsJSON, &alert.ReviewedBy,
		&alert.ReviewedAt, &alert.CreatedAt, &alert.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(detailsJSON, &alert.Details); err != nil {
		return nil, err
	}

	return &alert, nil
}

func (r *repository) UpdateAlertStatus(ctx context.Context, id uuid.UUID, status AlertStatus, reviewedBy uuid.UUID) error {
	query := `
		UPDATE fraud_alerts
		SET status = $1, reviewed_by = $2, reviewed_at = $3, updated_at = $4
		WHERE id = $5
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, status, reviewedBy, now, now, id)
	return err
}

func (r *repository) ListAlerts(ctx context.Context, status AlertStatus, limit, offset int) ([]FraudAlert, error) {
	query := `
		SELECT id, transaction_id, rule_id, risk_score, status, reason, details, reviewed_by, reviewed_at, created_at, updated_at
		FROM fraud_alerts
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []FraudAlert
	for rows.Next() {
		var alert FraudAlert
		var detailsJSON []byte

		err := rows.Scan(
			&alert.ID, &alert.TransactionID, &alert.RuleID, &alert.RiskScore,
			&alert.Status, &alert.Reason, &detailsJSON, &alert.ReviewedBy,
			&alert.ReviewedAt, &alert.CreatedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(detailsJSON, &alert.Details); err != nil {
			return nil, err
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

func (r *repository) CreateScore(ctx context.Context, score *FraudScore) error {
	featuresJSON, err := json.Marshal(score.Features)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO fraud_scores (id, transaction_id, account_id, amount_score, velocity_score, time_score, pattern_score, ml_score, total_score, features, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	score.ID = uuid.New()
	score.CreatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		score.ID, score.TransactionID, score.AccountID, score.AmountScore,
		score.VelocityScore, score.TimeScore, score.PatternScore, score.MLScore,
		score.TotalScore, featuresJSON, score.CreatedAt,
	)

	return err
}

func (r *repository) GetScoresByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]FraudScore, error) {
	query := `
		SELECT id, transaction_id, account_id, amount_score, velocity_score, time_score, pattern_score, ml_score, total_score, features, created_at
		FROM fraud_scores
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []FraudScore
	for rows.Next() {
		var score FraudScore
		var featuresJSON []byte

		err := rows.Scan(
			&score.ID, &score.TransactionID, &score.AccountID, &score.AmountScore,
			&score.VelocityScore, &score.TimeScore, &score.PatternScore, &score.MLScore,
			&score.TotalScore, &featuresJSON, &score.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(featuresJSON, &score.Features); err != nil {
			return nil, err
		}

		scores = append(scores, score)
	}

	return scores, rows.Err()
}

func (r *repository) GetRecentTransactions(ctx context.Context, accountID uuid.UUID, since time.Time) ([]RecentTransaction, error) {
	query := `
		SELECT amount, type, created_at
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
		  AND created_at >= $2
		  AND status = 'COMPLETED'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []RecentTransaction
	for rows.Next() {
		var tx RecentTransaction
		err := rows.Scan(&tx.Amount, &tx.Type, &tx.Timestamp)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (r *repository) GetTransactionCount(ctx context.Context, accountID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
		  AND created_at >= $2
		  AND status = 'COMPLETED'
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, accountID, since).Scan(&count)
	return count, err
}
