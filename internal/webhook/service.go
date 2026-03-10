package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/real-time-payments/pkg/logger"
)

type Service interface {
	// Subscriptions
	CreateSubscription(ctx context.Context, userID uuid.UUID, req *CreateSubscriptionRequest) (*Subscription, error)
	GetSubscriptionsByUser(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req *UpdateSubscriptionRequest) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error

	// Event Trigger (MAIN METHOD)
	TriggerEvent(ctx context.Context, eventType string, eventID uuid.UUID, payload map[string]interface{}) error

	// Delivery Management
	GetDeliveryHistory(ctx context.Context, subID uuid.UUID, limit, offset int) ([]Delivery, error)
	RetryDelivery(ctx context.Context, deliveryID uuid.UUID) error
}

type service struct {
	repo       Repository
	httpClient *http.Client
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *service) CreateSubscription(ctx context.Context, userID uuid.UUID, req *CreateSubscriptionRequest) (*Subscription, error) {
	// Generate secret
	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	sub := &Subscription{
		UserID:         userID,
		URL:            req.URL,
		Secret:         secret,
		Events:         req.Events,
		IsActive:       true,
		RetryCount:     req.RetryCount,
		TimeoutSeconds: req.TimeoutSeconds,
		Description:    req.Description,
	}

	if sub.RetryCount == 0 {
		sub.RetryCount = 3 // default
	}
	if sub.TimeoutSeconds == 0 {
		sub.TimeoutSeconds = 30 // default
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	logger.Info().
		Str("subscription_id", sub.ID.String()).
		Str("user_id", userID.String()).
		Str("url", sub.URL).
		Msg("webhook subscription created")

	return sub, nil
}

func (s *service) GetSubscriptionsByUser(ctx context.Context, userID uuid.UUID) ([]Subscription, error) {
	return s.repo.GetSubscriptionsByUserID(ctx, userID)
}

func (s *service) UpdateSubscription(ctx context.Context, id uuid.UUID, req *UpdateSubscriptionRequest) error {
	updates := make(map[string]interface{})

	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.RetryCount != nil {
		updates["retry_count"] = *req.RetryCount
	}
	if req.TimeoutSeconds != nil {
		updates["timeout_seconds"] = *req.TimeoutSeconds
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	return s.repo.UpdateSubscription(ctx, id, updates)
}

func (s *service) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteSubscription(ctx, id)
}

// TriggerEvent - MAIN METHOD that triggers webhooks
func (s *service) TriggerEvent(ctx context.Context, eventType string, eventID uuid.UUID, payload map[string]interface{}) error {
	// Get all active subscriptions for this event
	subs, err := s.repo.GetActiveSubscriptionsByEvent(ctx, EventType(eventType))
	if err != nil {
		logger.Error().Err(err).Str("event", eventType).Msg("failed to get subscriptions")
		return err
	}

	if len(subs) == 0 {
		logger.Debug().Str("event", eventType).Msg("no active subscriptions for event")
		return nil
	}

	// Create deliveries for each subscription
	for _, sub := range subs {
		delivery := &Delivery{
			SubscriptionID: sub.ID,
			EventType:      EventType(eventType),
			EventID:        eventID,
			Payload:        payload,
			Status:         DeliveryStatusPending,
			AttemptCount:   0,
			MaxAttempts:    sub.RetryCount,
		}

		if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
			logger.Error().Err(err).Msg("failed to create delivery")
			continue
		}

		// Send webhook asynchronously
		go s.sendWebhook(delivery, &sub)
	}

	logger.Info().
		Str("event", eventType).
		Str("event_id", eventID.String()).
		Int("subscribers", len(subs)).
		Msg("webhook event triggered")

	return nil
}

func (s *service) sendWebhook(delivery *Delivery, sub *Subscription) {
	ctx := context.Background()

	// Prepare payload
	webhookPayload := WebhookPayload{
		Event:     delivery.EventType,
		EventID:   delivery.EventID,
		Data:      delivery.Payload,
		Timestamp: time.Now(),
	}

	// Create signature
	payloadBytes, _ := json.Marshal(webhookPayload.Data)
	signature := generateSignature(payloadBytes, sub.Secret)
	webhookPayload.Signature = signature

	// Marshal full payload
	body, err := json.Marshal(webhookPayload)
	if err != nil {
		s.markDeliveryFailed(ctx, delivery.ID, 0, "failed to marshal payload", "")
		return
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", sub.URL, bytes.NewBuffer(body))
	if err != nil {
		s.markDeliveryFailed(ctx, delivery.ID, 0, "failed to create request", "")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", string(delivery.EventType))
	req.Header.Set("X-Webhook-ID", delivery.ID.String())

	// Send
	delivery.AttemptCount++
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("delivery_id", delivery.ID.String()).
			Str("url", sub.URL).
			Msg("webhook delivery failed, scheduling retry")
		s.scheduleRetry(ctx, delivery)
		return
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	// Check status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success!
		s.markDeliverySuccess(ctx, delivery.ID, resp.StatusCode, string(responseBody))
	} else {
		logger.Warn().
			Int("status_code", resp.StatusCode).
			Str("delivery_id", delivery.ID.String()).
			Msg("webhook delivery received non-2xx status, scheduling retry")
		s.scheduleRetry(ctx, delivery)
	}
}

func (s *service) scheduleRetry(ctx context.Context, delivery *Delivery) {
	if delivery.AttemptCount >= delivery.MaxAttempts {
		// Max attempts reached, mark as failed
		s.markDeliveryFailed(ctx, delivery.ID, delivery.AttemptCount, "max attempts reached", "")
		return
	}

	// Exponential backoff: 1min, 5min, 30min
	retryDelays := []time.Duration{1 * time.Minute, 5 * time.Minute, 30 * time.Minute}
	var delay time.Duration
	if delivery.AttemptCount-1 < len(retryDelays) {
		delay = retryDelays[delivery.AttemptCount-1]
	} else {
		delay = 1 * time.Hour
	}

	nextRetry := time.Now().Add(delay)

	updates := map[string]interface{}{
		"status":        string(DeliveryStatusRetrying),
		"attempt_count": delivery.AttemptCount,
		"next_retry_at": nextRetry,
	}

	s.repo.UpdateDelivery(ctx, delivery.ID, updates)

	logger.Warn().
		Str("delivery_id", delivery.ID.String()).
		Int("attempt", delivery.AttemptCount).
		Time("next_retry", nextRetry).
		Msg("webhook delivery scheduled for retry")
}

func (s *service) markDeliverySuccess(ctx context.Context, deliveryID uuid.UUID, statusCode int, responseBody string) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":               string(DeliveryStatusSuccess),
		"response_status_code": statusCode,
		"response_body":        responseBody,
		"delivered_at":         now,
	}

	s.repo.UpdateDelivery(ctx, deliveryID, updates)

	logger.Info().
		Str("delivery_id", deliveryID.String()).
		Int("status_code", statusCode).
		Msg("webhook delivered successfully")
}

func (s *service) markDeliveryFailed(ctx context.Context, deliveryID uuid.UUID, attemptCount int, errorMsg, responseBody string) {
	updates := map[string]interface{}{
		"status":        string(DeliveryStatusFailed),
		"attempt_count": attemptCount,
		"error_message": errorMsg,
		"response_body": responseBody,
	}

	s.repo.UpdateDelivery(ctx, deliveryID, updates)

	logger.Error().
		Str("delivery_id", deliveryID.String()).
		Str("error", errorMsg).
		Msg("webhook delivery failed permanently")
}

func (s *service) GetDeliveryHistory(ctx context.Context, subID uuid.UUID, limit, offset int) ([]Delivery, error) {
	return s.repo.GetDeliveriesBySubscription(ctx, subID, limit, offset)
}

func (s *service) RetryDelivery(ctx context.Context, deliveryID uuid.UUID) error {
	delivery, err := s.repo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		return err
	}

	sub, err := s.repo.GetSubscriptionByID(ctx, delivery.SubscriptionID)
	if err != nil {
		return err
	}

	go s.sendWebhook(delivery, sub)
	return nil
}

// Helper functions
func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
