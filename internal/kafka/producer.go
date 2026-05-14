package kafka

import (
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
)

var (
	AsyncProducer sarama.AsyncProducer
)

func InitProducer(brokers []string) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Flush.Messages = 500
	config.Producer.Flush.Frequency = 50 * time.Millisecond

	// Use a more resilient approach for producer initialization
	var err error
	AsyncProducer, err = sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error creating Kafka async producer: %v", err)
	}

	// Listen for errors in a background goroutine
	go func() {
		for err := range AsyncProducer.Errors() {
			log.Printf("Kafka producer error: %v", err)
			// Here we could implement a local fallback as mentioned in Phase 1 instructions
			SaveToFallback(err.Msg)
		}
	}()
}

func SaveToFallback(msg *sarama.ProducerMessage) {
	// Simple local fallback: append to a file
	f, err := os.OpenFile("kafka_fallback.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open fallback log: %v", err)
		return
	}
	defer f.Close()

	val, _ := msg.Value.Encode()
	logLine := time.Now().Format(time.RFC3339) + " | Topic: " + msg.Topic + " | Value: " + string(val) + "\n"
	if _, err := f.WriteString(logLine); err != nil {
		log.Printf("Failed to write to fallback log: %v", err)
	}
}
