package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"lendogo-backend/internal/services" // Update with your actual module path

	"github.com/segmentio/kafka-go"
)

type PaymentConsumer struct {
	Reader      *kafka.Reader
	LoanService services.LoanService // 👈 Inject the service it needs to talk to!
}

// Pass the service into the constructor
func NewPaymentConsumer(brokerURL string, topic string, groupID string, loanService services.LoanService) *PaymentConsumer {
	return &PaymentConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{brokerURL},
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		LoanService: loanService,
	}
}

func (c *PaymentConsumer) Start(ctx context.Context) {
	log.Printf("📥 Kafka Consumer listening on topic: %s", c.Reader.Config().Topic)

	for {
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Kafka Consumer Error: %v", err)
			continue // Don't crash the loop, just keep trying
		}

		// 1. Unmarshal the standard envelope we created in utils/kafka.go
		var envelope map[string]interface{}
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			log.Printf("❌ Failed to parse Kafka message: %v", err)
			continue
		}

		// Ensure event_type exists and is a string
		eventType, ok := envelope["event_type"].(string)
		if !ok {
			log.Println("❌ Kafka message missing valid 'event_type'")
			continue
		}

		// 2. Route the event to the correct business logic!
		switch eventType {

		// 👇 NEW CASE: Catching the System Recharge Event! 👇
		case "SYSTEM_RECHARGE_SUCCESS":
			fmt.Println("✅ [KAFKA CONSUMER CAUGHT EVENT] -> SYSTEM_RECHARGE_SUCCESS")
			
			// Safely extract the data payload (Amount, Transaction ID, etc.)
			if data, ok := envelope["data"].(map[string]interface{}); ok {
				fmt.Printf("💰 System Wallet Recharged by: ₹%v (Txn: %v)\n", data["amount"], data["transaction_id"])
			}
			
			// TODO Future: Save this to an Audit Ledger database table
			// TODO Future: Send an email to the Super Admin confirming the top-up

		case "PAYMENT_SUCCESS":
			fmt.Println("✅ [KAFKA EVENT RECOGNIZED] -> PAYMENT_SUCCESS")
			// TODO: c.LoanService.MarkEMIPaid(envelope["data"])

		case "LOAN_DEFAULTED":
			fmt.Println("[KAFKA EVENT RECOGNIZED] -> LOAN_DEFAULTED")
			// TODO: c.LoanService.ApplyPenalty(...)
		}
	}
}