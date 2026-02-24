// Example: structured fields and the Logger interface for dependency injection.
//
// Run: go run ./examples/fields
package main

import "github.com/jenish-rudani/logsift"

func main() {
	logsift.SetLevel("debug")
	logsift.SetFormat("forceColor")

	// --- With: attach a single field ---
	reqLogger := logsift.With("request_id", "req-abc-123")
	reqLogger.Info("handling request")
	reqLogger.Debug("parsing body")

	// --- WithFields: attach multiple fields ---
	userLogger := logsift.WithFields(map[string]interface{}{
		"user_id":  42,
		"username": "alice",
	})
	userLogger.Info("user authenticated")
	userLogger.Warn("password expires soon")

	// --- Chaining: fields accumulate ---
	detailed := logsift.With("service", "billing").With("region", "us-east-1")
	detailed.Infof("invoice generated: $%.2f", 99.95)

	// --- Logger interface for dependency injection ---
	svc := NewOrderService(logsift.With("component", "orders"))
	svc.PlaceOrder("item-789", 3)
}

// OrderService shows how to accept logsift.Logger as a dependency.
type OrderService struct {
	log logsift.Logger
}

func NewOrderService(log logsift.Logger) *OrderService {
	return &OrderService{log: log}
}

func (s *OrderService) PlaceOrder(itemID string, qty int) {
	s.log.Infof("placing order: item=%s qty=%d", itemID, qty)

	// The injected logger carries its fields through
	s.log.With("item_id", itemID).Debug("inventory check passed")
	s.log.Info("order confirmed")
}
