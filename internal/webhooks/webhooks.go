// Package webhooks provides webhook functionality for Fieldstone.
// Supports event-driven callbacks with retry logic, HMAC signatures, and dead letter queue.
//
// Features:
// - Event filtering (specific collections/actions)
// - HMAC-SHA256 signature verification
// - Exponential backoff retry
// - Dead letter queue for failed webhooks
// - Webhook management API
// - Request/response logging
//
// Example webhook payload:
// {
//   "event": "records.create",
//   "timestamp": "2026-01-15T10:30:00Z",
//   "data": { ...record data... },
//   "signature": "sha256=abc123..."
// }
package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Webhook represents a configured webhook endpoint
type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Secret    string   `json:"secret"` // For HMAC signature
	Events    []string `json:"events"` // Filter: which events to send
	Active    bool     `json:"active"`
	Retries   int      `json:"retries"` // Max retry attempts
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Event represents a webhook event
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`      // e.g., "records.create"
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Signature string          `json:"signature"`
}

// Delivery represents a webhook delivery attempt
type Delivery struct {
	ID         string    `json:"id"`
	WebhookID  string    `json:"webhookId"`
	EventID    string    `json:"eventId"`
	Status     string    `json:"status"` // pending, delivered, failed
	Attempts   int       `json:"attempts"`
	Response   string    `json:"response"`
	StatusCode int       `json:"statusCode"`
	CreatedAt  time.Time `json:"createdAt"`
	NextRetry  time.Time `json:"nextRetry"`
}

// Manager manages webhooks and deliveries
type Manager struct {
	store     WebhookStore
	delivery  DeliveryStore
	http      *http.Client
	hookQueue chan *Event
	quit      chan struct{}
}

// WebhookStore defines storage interface for webhooks
type WebhookStore interface {
	Create(ctx context.Context, webhook *Webhook) error
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Webhook, error)
	List(ctx context.Context) ([]*Webhook, error)
	ListByEvent(ctx context.Context, event string) ([]*Webhook, error)
}

// DeliveryStore defines storage interface for deliveries
type DeliveryStore interface {
	Create(ctx context.Context, delivery *Delivery) error
	Update(ctx context.Context, delivery *Delivery) error
	Get(ctx context.Context, id string) (*Delivery, error)
	ListPending(ctx context.Context, limit int) ([]*Delivery, error)
	ListByWebhook(ctx context.Context, webhookID string, limit int) ([]*Delivery, error)
}

// NewManager creates a new webhook manager
func NewManager(store WebhookStore, delivery DeliveryStore) *Manager {
	return &Manager{
		store:     store,
		delivery:  delivery,
		http:      &http.Client{Timeout: 30 * time.Second},
		hookQueue: make(chan *Event, 1000),
		quit:      make(chan struct{}),
	}
}

// Start begins processing webhook events
func (m *Manager) Start(workers int) {
	for i := 0; i  workers; i++ {
		go m.worker()
	}
}

// Stop gracefully shuts down webhook processing
func (m *Manager) Stop() {
	close(m.quit)
}

// worker processes webhook events
func (m *Manager) worker() {
	for {
		select {
		case event := range m.hookQueue:
			m.processEvent(event)
		case m.quit:
			return
		}
	}
}

// Trigger sends an event to all matching webhooks
func (m *Manager) Trigger(ctx context.Context, eventType string, data interface{}) error {
	// Marshal data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	
	// Find matching webhooks
	webhooks, err := m.store.ListByEvent(ctx, eventType)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}
	
	// Create event
	event := &Event{
		ID:        generateID(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      json.RawMessage(jsonData),
	}
	
	// Queue for each webhook
	for _, webhook := range webhooks {
		if !webhook.Active {
			continue
		}
		
		// Create delivery record
		delivery := &Delivery{
			ID:        generateID(),
			WebhookID: webhook.ID,
			EventID:   event.ID,
			Status:    "pending",
			Attempts:  0,
			CreatedAt: time.Now().UTC(),
		}
		
		if err := m.delivery.Create(ctx, delivery); err != nil {
			log.Error().Err(err).Str("webhook", webhook.ID).Msg("Failed to create delivery")
			continue
		}
		
		// Queue event
		select {
		case m.hookQueue struct{}{}event:
		default:
			log.Warn().Str("webhook", webhook.ID).Msg("Webhook queue full, dropping event")
		}
	}
	
	return nil
}

// processEvent delivers event to webhook
func (m *Manager) processEvent(event *Event) {
	ctx := context.Background()
	
	// Get delivery
	deliveries, err := m.delivery.ListPending(ctx, 100)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list pending deliveries")
		return
	}
	
	for _, delivery := range deliveries {
		// Get webhook
		webhook, err := m.store.Get(ctx, delivery.WebhookID)
		if err != nil {
			log.Error().Err(err).Str("webhook", delivery.WebhookID).Msg("Webhook not found")
			continue
		}
		
		// Attempt delivery
		if err := m.deliver(ctx, webhook, event, delivery); err != nil {
			// Retry with exponential backoff
			delivery.Attempts++
			if delivery.Attempts  webhook.Retries {
				delivery.Status = "failed"
				m.delivery.Update(ctx, delivery)
				log.Error().
					Err(err).
					Str("webhook", webhook.ID).
					Int("attempts", delivery.Attempts).
					Msg("Webhook delivery failed after retries")
			} else {
				// Schedule retry
				backoff := time.Duration(1delivery.Attempts) * time.Second
				delivery.NextRetry = time.Now().Add(backoff)
				m.delivery.Update(ctx, delivery)
				log.Warn().
					Err(err).
					Str("webhook", webhook.ID).
					Int("attempt", delivery.Attempts).
					Dur("retry_after", backoff).
					Msg("Webhook delivery failed, scheduling retry")
			}
		}
	}
}

// deliver sends HTTP request to webhook
func (m *Manager) deliver(ctx context.Context, webhook *Webhook, event *Event, delivery *Delivery) error {
	// Generate signature
	event.Signature = generateSignature(event.Data, webhook.Secret)
	
	// Marshal event
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-ID", webhook.ID)
	req.Header.Set("X-Event-ID", event.ID)
	req.Header.Set("X-Signature", event.Signature)
	
	// Send request
	resp, err := m.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, _ := io.ReadAll(resp.Body)
	delivery.Response = string(body)
	delivery.StatusCode = resp.StatusCode
	
	// Check status
	if resp.StatusCode >= 200 && resp.StatusCode  300 {
		delivery.Status = "delivered"
		m.delivery.Update(ctx, delivery)
		log.Info().
			Str("webhook", webhook.ID).
			Str("event", event.Type).
			Int("status", resp.StatusCode).
			Msg("Webhook delivered successfully")
		return nil
	}
	
	return fmt.Errorf("webhook returned status %d", resp.StatusCode)
}

// generateSignature creates HMAC-SHA256 signature
func generateSignature(data []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies webhook signature
func VerifySignature(data []byte, signature string, secret string) bool {
	expected := generateSignature(data, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// generateID generates unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Event types
const (
	EventRecordsCreate = "records.create"
	EventRecordsUpdate = "records.update"
	EventRecordsDelete = "records.delete"
	EventUsersCreate   = "users.create"
	EventUsersUpdate   = "users.update"
	EventUsersDelete   = "users.delete"
	EventAuthLogin     = "auth.login"
	EventAuthLogout    = "auth.logout"
)
