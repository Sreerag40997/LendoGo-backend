package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/google/uuid"

	"lendogo-backend/internal/services"
)

type PaymentConsumer struct {
	Reader              *kafka.Reader
	LoanService         services.LoanService
	NotificationService services.NotificationService // 👈 ADDED: For sending payment alerts
}

// 👇 UPDATED: Constructor now requires the NotificationService
func NewPaymentConsumer(brokerURL string, topic string, groupID string, loanService services.LoanService, notifService services.NotificationService) *PaymentConsumer {
	return &PaymentConsumer{
		Reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{brokerURL},
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		LoanService:         loanService,
		NotificationService: notifService,
	}
}

func (c *PaymentConsumer) Start(ctx context.Context) {
	log.Printf("📥 Payment Consumer listening on topic: %s", c.Reader.Config().Topic)

	for {
		msg, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Kafka Consumer Error: %v", err)
			continue
		}

		var envelope map[string]interface{}
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			log.Printf("❌ Failed to parse Kafka message: %v", err)
			continue
		}

		eventType, ok := envelope["event_type"].(string)
		if !ok {
			log.Println("❌ Kafka message missing valid 'event_type'")
			continue
		}

		switch eventType {

		// ==========================================
		// 1. ADMIN SYSTEM WALLET TOP-UP
		// ==========================================
		case "SYSTEM_RECHARGE_SUCCESS":
			fmt.Println("✅ [KAFKA EVENT] -> SYSTEM_RECHARGE_SUCCESS")
			
			if data, ok := envelope["data"].(map[string]interface{}); ok {
				fmt.Printf("💰 System Wallet Recharged by: ₹%v (Txn: %v)\n", data["amount"], data["transaction_id"])
			}

		// ==========================================
		// 2. USER PAYS AN EMI 
		// ==========================================
		case "PAYMENT_SUCCESS":
			fmt.Println("✅ [KAFKA EVENT] -> PAYMENT_SUCCESS")
			
			data := envelope["data"].(map[string]interface{})
			userIDStr := fmt.Sprintf("%v", data["user_id"])
			scheduleIDStr := fmt.Sprintf("%v", data["schedule_id"])
			amount := data["amount"]

			fmt.Printf("💸 Processing EMI Payment for Schedule %v\n", scheduleIDStr)

			// Step A: Update the EMI Schedule row to 'Paid'
			err := c.LoanService.MarkEMIPaid(context.Background(), scheduleIDStr)
			if err != nil {
				log.Printf("❌ Failed to update EMI status: %v", err)
			} else {
				fmt.Println("✅ EMI marked as PAID in database!")
			}

			// Step B: Push In-App Notification
			userID, parseErr := uuid.Parse(userIDStr)
			if parseErr == nil {
				c.NotificationService.SendNotification(userID, fmt.Sprintf("Your EMI payment of ₹%v was successful. Thank you!", amount), "SUCCESS", "payment")
			}

		// ==========================================
		// 3. CRON JOB DETECTS MISSED EMI
		// ==========================================
		case "LOAN_DEFAULTED":
			fmt.Println("⚠️ [KAFKA EVENT] -> LOAN_DEFAULTED")
			
			data := envelope["data"].(map[string]interface{})
			userIDStr := fmt.Sprintf("%v", data["user_id"])
			loanIDStr := fmt.Sprintf("%v", data["loan_id"])

			// Step A: Apply late fee penalty to the loan balance
			err := c.LoanService.ApplyPenalty(context.Background(), loanIDStr)
			if err != nil {
				log.Printf("❌ Failed to apply penalty: %v", err)
			}

			// Step B: Push In-App Warning Notification
			userID, parseErr := uuid.Parse(userIDStr)
			if parseErr == nil {
				c.NotificationService.SendNotification(userID, "Your loan EMI is overdue. A late penalty has been applied to your balance.", "WARNING", "loan")
			}
		}
	}
}