package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/google/uuid"

	"lendogo-backend/internal/services"
	"lendogo-backend/utils"
)

type LoanConsumer struct {
	Reader              *kafka.Reader
	LoanService         services.LoanService
	NotificationService services.NotificationService
}

func NewLoanConsumer(brokerURL string, topic string, groupID string, loanService services.LoanService, notifService services.NotificationService) *LoanConsumer {
	return &LoanConsumer{
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

func (c *LoanConsumer) Start(ctx context.Context) {
	log.Printf("📥 Loan Consumer listening on topic: %s", c.Reader.Config().Topic)

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
			continue
		}

		switch eventType {
		case "LOAN_DISBURSED":
			fmt.Println("✅ [KAFKA CAUGHT EVENT] -> LOAN_DISBURSED")

			payload := envelope["data"].(map[string]interface{})
			loanID := fmt.Sprintf("%v", payload["loan_id"])
			userID := fmt.Sprintf("%v", payload["user_id"])
			amount := payload["amount"]

			fmt.Printf("💸 Processing Background Tasks for Loan %v (User: %v, Amt: ₹%v)\n", loanID, userID, amount)

			// 1. GENERATE EMI SCHEDULE
			err := c.LoanService.GenerateEMISchedule(context.Background(), loanID)
			if err != nil {
				log.Printf("❌ Failed to generate EMI schedule: %v", err)
			} else {
				fmt.Println("✅ EMI Schedule Generated & Loan History Updated!")
			}

			// 2. SEND IN-APP NOTIFICATION
			importUUID, err := uuid.Parse(userID)
			if err == nil {
				c.NotificationService.SendNotification(importUUID, fmt.Sprintf("Your loan of ₹%v has been disbursed to your wallet!", amount), "SUCCESS", "loan")
			}

			// 3. SEND EMAIL (Non-blocking)
			go func(uID string, amt interface{}) {
				utils.SendDisbursalEmail(uID, amt)
			}(userID, amount)

			fmt.Println("🏁 All background tasks for LOAN_DISBURSED completed successfully!")
		}
	}
}