package webhook

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository interface {
	// Subscriptions
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	GetActiveSubscriptionsByEvent(ctx context.Context, eventType EventType) ([]Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error

	// Deliveries
	CreateDelivery(ctx context.Context, delivery *Delivery) error
	GetDeliveryByID(ctx context.Context, id uuid.UUID) (*Delivery, error)
	GetDeliveriesBySubscription(ctx context.Context, subID uuid.UUID, limit, offset int) ([]Delivery, error)
	GetPendingRetries(ctx context.Context, limit int) ([]Delivery, error)
	UpdateDelivery(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	query := `
		INSERT INTO webhook_subscriptions
		(id, user_id, url, secret, events, is_active, retry_count, timeout_seconds, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	// Convert events to string array
	events := make([]string, len(sub.Events))
	for i, e := range sub.Events {
		events[i] = string(e)
	}

	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.UserID, sub.URL, sub.Secret, pq.Array(events),
		sub.IsActive, sub.RetryCount, sub.TimeoutSeconds, sub.Description,
		sub.CreatedAt, sub.UpdatedAt,
	)

	return err
}

func (r *repository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	query := `
		SELECT id, user_id, url, secret, events, is_active, retry_count, timeout_seconds, description, created_at, updated_at
		FROM webhook_subscriptions
		WHERE id = $1
	`

	var sub Subscription
	var events pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.URL, &sub.Secret, &events,
		&sub.IsActive, &sub.RetryCount, &sub.TimeoutSeconds, &sub.Description,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert events
	sub.Events = make([]EventType, len(events))
	for i, e := range events {
		sub.Events[i] = EventType(e)
	}

	return &sub, nil
}

func (r *repository) GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	query := `
		SELECT id, user_id, url, secret, events, is_active, retry_count, timeout_seconds, description, created_at, updated_at
		FROM webhook_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription
		var events pq.StringArray

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.URL, &sub.Secret, &events,
			&sub.IsActive, &sub.RetryCount, &sub.TimeoutSeconds, &sub.Description,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		sub.Events = make([]EventType, len(events))
		for i, e := range events {
			sub.Events[i] = EventType(e)
		}

		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

func (r *repository) GetActiveSubscriptionsByEvent(ctx context.Context, eventType EventType) ([]Subscription, error) {
	query := `
		SELECT id, user_id, url, secret, events, is_active, retry_count, timeout_seconds, description, created_at, updated_at
		FROM webhook_subscriptions
		WHERE is_active = true AND $1 = ANY(events)
	`

	rows, err := r.db.QueryContext(ctx, query, string(eventType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription
		var events pq.StringArray

		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.URL, &sub.Secret, &events,
			&sub.IsActive, &sub.RetryCount, &sub.TimeoutSeconds, &sub.Description,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		sub.Events = make([]EventType, len(events))
		for i, e := range events {
			sub.Events[i] = EventType(e)
		}

		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

func (r *repository) UpdateSubscription(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE webhook_subscriptions SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		query += ", " + key + " = $" + string(rune(argIndex+'0'))
		args = append(args, value)
		argIndex++
	}

	query += " WHERE id = $" + string(rune(argIndex+'0'))
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM webhook_subscriptions WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *repository) CreateDelivery(ctx context.Context, delivery *Delivery) error {
	payloadJSON, err := json.Marshal(delivery.Payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO webhook_deliveries
		(id, subscription_id, event_type, event_id, payload, status, attempt_count, max_attempts, next_retry_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	delivery.ID = uuid.New()
	delivery.CreatedAt = time.Now()
	delivery.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		delivery.ID, delivery.SubscriptionID, delivery.EventType, delivery.EventID,
		payloadJSON, delivery.Status, delivery.AttemptCount, delivery.MaxAttempts,
		delivery.NextRetryAt, delivery.CreatedAt, delivery.UpdatedAt,
	)

	return err
}

func (r *repository) GetDeliveryByID(ctx context.Context, id uuid.UUID) (*Delivery, error) {
	query := `
		SELECT id, subscription_id, event_type, event_id, payload, status, attempt_count, max_attempts,
		       next_retry_at, response_status_code, response_body, error_message, delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE id = $1
	`

	var delivery Delivery
	var payloadJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&delivery.ID, &delivery.SubscriptionID, &delivery.EventType, &delivery.EventID,
		&payloadJSON, &delivery.Status, &delivery.AttemptCount, &delivery.MaxAttempts,
		&delivery.NextRetryAt, &delivery.ResponseStatusCode, &delivery.ResponseBody,
		&delivery.ErrorMessage, &delivery.DeliveredAt, &delivery.CreatedAt, &delivery.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payloadJSON, &delivery.Payload); err != nil {
		return nil, err
	}

	return &delivery, nil
}

func (r *repository) GetDeliveriesBySubscription(ctx context.Context, subID uuid.UUID, limit, offset int) ([]Delivery, error) {
	query := `
		SELECT id, subscription_id, event_type, event_id, payload, status, attempt_count, max_attempts,
		       next_retry_at, response_status_code, response_body, error_message, delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, subID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		var delivery Delivery
		var payloadJSON []byte

		err := rows.Scan(
			&delivery.ID, &delivery.SubscriptionID, &delivery.EventType, &delivery.EventID,
			&payloadJSON, &delivery.Status, &delivery.AttemptCount, &delivery.MaxAttempts,
			&delivery.NextRetryAt, &delivery.ResponseStatusCode, &delivery.ResponseBody,
			&delivery.ErrorMessage, &delivery.DeliveredAt, &delivery.CreatedAt, &delivery.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(payloadJSON, &delivery.Payload); err != nil {
			return nil, err
		}

		deliveries = append(deliveries, delivery)
	}

	return deliveries, rows.Err()
}

func (r *repository) GetPendingRetries(ctx context.Context, limit int) ([]Delivery, error) {
	query := `
		SELECT id, subscription_id, event_type, event_id, payload, status, attempt_count, max_attempts,
		       next_retry_at, response_status_code, response_body, error_message, delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE status = 'RETRYING' AND next_retry_at <= NOW()
		ORDER BY next_retry_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		var delivery Delivery
		var payloadJSON []byte

		err := rows.Scan(
			&delivery.ID, &delivery.SubscriptionID, &delivery.EventType, &delivery.EventID,
			&payloadJSON, &delivery.Status, &delivery.AttemptCount, &delivery.MaxAttempts,
			&delivery.NextRetryAt, &delivery.ResponseStatusCode, &delivery.ResponseBody,
			&delivery.ErrorMessage, &delivery.DeliveredAt, &delivery.CreatedAt, &delivery.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(payloadJSON, &delivery.Payload); err != nil {
			return nil, err
		}

		deliveries = append(deliveries, delivery)
	}

	return deliveries, rows.Err()
}

func (r *repository) UpdateDelivery(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE webhook_deliveries SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		query += ", " + key + " = $" + string(rune(argIndex+'0'))
		args = append(args, value)
		argIndex++
	}

	query += " WHERE id = $" + string(rune(argIndex+'0'))
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
