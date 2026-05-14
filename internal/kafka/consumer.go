package kafka

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/IBM/sarama"
)

type ConsumerHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

type ConsumerGroupHandler struct {
	Ready   chan bool
	Handler ConsumerHandler
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.Ready)
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := h.processWithRetry(session.Context(), msg)
		if err != nil {
			log.Printf("Failed to process message after retries: %v", err)
			h.sendToDLQ(msg, err)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *ConsumerGroupHandler) processWithRetry(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		lastErr = h.Handler(ctx, msg)
		if lastErr == nil {
			return nil
		}
		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Printf("Retry %d/3 for message on topic %s after %v: %v", i+1, msg.Topic, backoff, lastErr)
		time.Sleep(backoff)
	}
	return lastErr
}

func (h *ConsumerGroupHandler) sendToDLQ(msg *sarama.ConsumerMessage, processErr error) {
	dlqTopic := fmt.Sprintf("%s.dlq", msg.Topic)
	dlqMsg := &sarama.ProducerMessage{
		Topic: dlqTopic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(msg.Value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("error"), Value: []byte(processErr.Error())},
			{Key: []byte("original_topic"), Value: []byte(msg.Topic)},
			{Key: []byte("failed_at"), Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	_, _, err := syncProducer.SendMessage(dlqMsg)
	if err != nil {
		log.Printf("Failed to send to DLQ %s: %v", dlqTopic, err)
	}
}

var syncProducer sarama.SyncProducer

func InitSyncProducer(brokers []string) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	var err error
	syncProducer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka sync producer: %v", err)
	}
}
