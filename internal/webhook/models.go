package webhook

import (
	"time"

	"github.com/google/uuid"
)

// EventType representa o tipo de evento
type EventType string

const (
	EventTransactionCreated   EventType = "transaction.created"
	EventTransactionCompleted EventType = "transaction.completed"
	EventTransactionFailed    EventType = "transaction.failed"
	EventFraudDetected        EventType = "fraud.detected"
	EventFraudConfirmed       EventType = "fraud.confirmed"
	EventAccountCreated       EventType = "account.created"
	EventAccountBalanceLow    EventType = "account.balance_low"
)

// DeliveryStatus representa o status de entrega do webhook
type DeliveryStatus string

const (
	DeliveryStatusPending  DeliveryStatus = "PENDING"
	DeliveryStatusSuccess  DeliveryStatus = "SUCCESS"
	DeliveryStatusFailed   DeliveryStatus = "FAILED"
	DeliveryStatusRetrying DeliveryStatus = "RETRYING"
)

// Subscription representa uma assinatura de webhook
type Subscription struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	URL            string      `json:"url"`
	Secret         string      `json:"secret"`
	Events         []EventType `json:"events"`
	IsActive       bool        `json:"is_active"`
	RetryCount     int         `json:"retry_count"`
	TimeoutSeconds int         `json:"timeout_seconds"`
	Description    string      `json:"description"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// Delivery representa uma tentativa de entrega de webhook
type Delivery struct {
	ID                 uuid.UUID              `json:"id"`
	SubscriptionID     uuid.UUID              `json:"subscription_id"`
	EventType          EventType              `json:"event_type"`
	EventID            uuid.UUID              `json:"event_id"`
	Payload            map[string]interface{} `json:"payload"`
	Status             DeliveryStatus         `json:"status"`
	AttemptCount       int                    `json:"attempt_count"`
	MaxAttempts        int                    `json:"max_attempts"`
	NextRetryAt        *time.Time             `json:"next_retry_at,omitempty"`
	ResponseStatusCode *int                   `json:"response_status_code,omitempty"`
	ResponseBody       string                 `json:"response_body,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	DeliveredAt        *time.Time             `json:"delivered_at,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// Event representa um evento a ser enviado
type Event struct {
	Type      EventType              `json:"type"`
	ID        uuid.UUID              `json:"id"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// WebhookPayload é o payload enviado no webhook
type WebhookPayload struct {
	Event     EventType              `json:"event"`
	EventID   uuid.UUID              `json:"event_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Signature string                 `json:"signature"` // HMAC-SHA256
}

// CreateSubscriptionRequest
type CreateSubscriptionRequest struct {
	URL            string      `json:"url" binding:"required,url"`
	Events         []EventType `json:"events" binding:"required,min=1"`
	RetryCount     int         `json:"retry_count"`
	TimeoutSeconds int         `json:"timeout_seconds"`
	Description    string      `json:"description"`
}

// UpdateSubscriptionRequest
type UpdateSubscriptionRequest struct {
	URL            *string      `json:"url,omitempty"`
	Events         *[]EventType `json:"events,omitempty"`
	IsActive       *bool        `json:"is_active,omitempty"`
	RetryCount     *int         `json:"retry_count,omitempty"`
	TimeoutSeconds *int         `json:"timeout_seconds,omitempty"`
	Description    *string      `json:"description,omitempty"`
}
