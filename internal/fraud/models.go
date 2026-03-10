package fraud

import (
	"time"

	"github.com/google/uuid"
)

// RuleType representa o tipo de regra de detecção
type RuleType string

const (
	RuleTypeHighAmount  RuleType = "HIGH_AMOUNT"
	RuleTypeVelocity    RuleType = "VELOCITY"
	RuleTypeTimePattern RuleType = "TIME_PATTERN"
	RuleTypeLocation    RuleType = "LOCATION"
	RuleTypeCustom      RuleType = "CUSTOM"
)

// Action representa a ação a ser tomada
type Action string

const (
	ActionFlag   Action = "FLAG"
	ActionBlock  Action = "BLOCK"
	ActionReview Action = "REVIEW"
)

// AlertStatus representa o status de um alerta
type AlertStatus string

const (
	AlertStatusPending       AlertStatus = "PENDING"
	AlertStatusConfirmed     AlertStatus = "CONFIRMED"
	AlertStatusFalsePositive AlertStatus = "FALSE_POSITIVE"
	AlertStatusResolved      AlertStatus = "RESOLVED"
)

// FraudRule representa uma regra de detecção de fraude
type FraudRule struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    RuleType               `json:"rule_type"`
	Condition   map[string]interface{} `json:"condition"`
	RiskScore   int                    `json:"risk_score"`
	Action      Action                 `json:"action"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// FraudAlert representa um alerta de fraude
type FraudAlert struct {
	ID            uuid.UUID              `json:"id"`
	TransactionID uuid.UUID              `json:"transaction_id"`
	RuleID        *uuid.UUID             `json:"rule_id,omitempty"`
	RiskScore     int                    `json:"risk_score"`
	Status        AlertStatus            `json:"status"`
	Reason        string                 `json:"reason"`
	Details       map[string]interface{} `json:"details"`
	ReviewedBy    *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time             `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// FraudScore representa o score de fraude de uma transação
type FraudScore struct {
	ID             uuid.UUID              `json:"id"`
	TransactionID  uuid.UUID              `json:"transaction_id"`
	AccountID      uuid.UUID              `json:"account_id"`
	AmountScore    int                    `json:"amount_score"`
	VelocityScore  int                    `json:"velocity_score"`
	TimeScore      int                    `json:"time_score"`
	PatternScore   int                    `json:"pattern_score"`
	MLScore        int                    `json:"ml_score"`
	TotalScore     int                    `json:"total_score"`
	Features       map[string]interface{} `json:"features"`
	CreatedAt      time.Time              `json:"created_at"`
}

// TransactionContext contém informações contextuais de uma transação para análise
type TransactionContext struct {
	TransactionID      uuid.UUID
	AccountID          uuid.UUID
	Amount             float64
	Type               string
	Timestamp          time.Time
	RecentTransactions []RecentTransaction
}

// RecentTransaction representa uma transação recente
type RecentTransaction struct {
	Amount    float64
	Type      string
	Timestamp time.Time
}

// FraudCheckResult resultado da verificação de fraude
type FraudCheckResult struct {
	IsFraud   bool         `json:"is_fraud"`
	RiskScore int          `json:"risk_score"`
	Reasons   []string     `json:"reasons"`
	Action    Action       `json:"action"`
	AlertID   *uuid.UUID   `json:"alert_id,omitempty"`
	Rules     []FraudRule  `json:"triggered_rules"`
}
