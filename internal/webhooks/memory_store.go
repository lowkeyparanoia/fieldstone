package webhooks

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory implementation of WebhookStore.
type MemoryStore struct {
	mu       sync.RWMutex
	webhooks map[string]*Webhook
}

// NewMemoryStore creates a new in-memory webhook store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		webhooks: make(map[string]*Webhook),
	}
}

func (s *MemoryStore) Create(ctx context.Context, webhook *Webhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhooks[webhook.ID] = webhook
	return nil
}

func (s *MemoryStore) Update(ctx context.Context, webhook *Webhook) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhooks[webhook.ID] = webhook
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.webhooks, id)
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, id string) (*Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.webhooks[id]
	if !ok {
		return nil, context.Canceled // placeholder error
	}
	return w, nil
}

func (s *MemoryStore) List(ctx context.Context) ([]*Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Webhook
	for _, w := range s.webhooks {
		out = append(out, w)
	}
	return out, nil
}

func (s *MemoryStore) ListByEvent(ctx context.Context, event string) ([]*Webhook, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Webhook
	for _, w := range s.webhooks {
		if !w.Active {
			continue
		}
		for _, e := range w.Events {
			if e == event || e == "*" {
				out = append(out, w)
				break
			}
		}
	}
	return out, nil
}

// MemoryDeliveryStore is an in-memory implementation of DeliveryStore.
type MemoryDeliveryStore struct {
	mu         sync.RWMutex
	deliveries map[string]*Delivery
}

// NewMemoryDeliveryStore creates a new in-memory delivery store.
func NewMemoryDeliveryStore() *MemoryDeliveryStore {
	return &MemoryDeliveryStore{
		deliveries: make(map[string]*Delivery),
	}
}

func (s *MemoryDeliveryStore) Create(ctx context.Context, delivery *Delivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries[delivery.ID] = delivery
	return nil
}

func (s *MemoryDeliveryStore) Update(ctx context.Context, delivery *Delivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries[delivery.ID] = delivery
	return nil
}

func (s *MemoryDeliveryStore) Get(ctx context.Context, id string) (*Delivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deliveries[id]
	if !ok {
		return nil, context.Canceled
	}
	return d, nil
}

func (s *MemoryDeliveryStore) ListPending(ctx context.Context, limit int) ([]*Delivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Delivery
	for _, d := range s.deliveries {
		if d.Status == "pending" {
			out = append(out, d)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (s *MemoryDeliveryStore) ListByWebhook(ctx context.Context, webhookID string, limit int) ([]*Delivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Delivery
	for _, d := range s.deliveries {
		if d.WebhookID == webhookID {
			out = append(out, d)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
