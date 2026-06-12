package utils

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	Writer *kafka.Writer
}

func NewKafkaProducer(brokerURL string) *KafkaProducer {
	return &KafkaProducer{
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(brokerURL),
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			Async:        false,
		},
	}
}

func (p *KafkaProducer) PublishEvent(ctx context.Context, topic, eventType string, payload interface{}) error {
	envelope := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Unix(),
		"data":       payload,
	}

	jsonData, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	err = p.Writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: jsonData,
	})
	if err != nil {
		log.Printf("❌ Kafka publish failed to %s: %v", topic, err)
		return err
	}

	log.Printf("🚀 Kafka Event [%s] sent to topic: %s", eventType, topic)
	return nil
}

func (p *KafkaProducer) Close() error {
	return p.Writer.Close()
}