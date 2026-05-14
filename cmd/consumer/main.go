package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"adtracker/internal/kafka"
	"adtracker/internal/models"
	"adtracker/internal/storage"

	"github.com/IBM/sarama"
)

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	brokerList := strings.Split(brokers, ",")

	kafka.InitSyncProducer(brokerList)

	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}

	group := "adtracker-consumer-group"
	client, err := sarama.NewConsumerGroup(brokerList, group, config)
	if err != nil {
		log.Fatalf("Error creating consumer group client: %v", err)
	}

	store := storage.NewStorage()

	handler := &ConsumerHandler{
		Ready:       make(chan bool),
		store:       store,
		batchSize:   500,
		flushInterval: 5 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	topics := []string{"events.impressions", "events.clicks", "events.conversions"}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			handler.Ready = make(chan bool)
		}
	}()

	<-handler.Ready
	log.Println("Consumer up and running...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}

	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

type ConsumerHandler struct {
	Ready         chan bool
	store         *storage.Storage
	batchSize     int
	flushInterval time.Duration

	mu          sync.Mutex
	impressions []models.Impression
	clicks      []models.Click
	conversions []models.Conversion
	lastFlush   time.Time
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	h.lastFlush = time.Now()
	close(h.Ready)
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return h.flush(context.Background())
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.processMessage(msg)
		session.MarkMessage(msg, "")

		h.mu.Lock()
		if len(h.impressions)+len(h.clicks)+len(h.conversions) >= h.batchSize || time.Since(h.lastFlush) >= h.flushInterval {
			h.mu.Unlock()
			if err := h.flush(session.Context()); err != nil {
				log.Printf("Flush error: %v", err)
			}
		} else {
			h.mu.Unlock()
		}
	}
	return nil
}

func (h *ConsumerHandler) processMessage(msg *sarama.ConsumerMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch msg.Topic {
	case "events.impressions":
		var event models.Impression
		if err := json.Unmarshal(msg.Value, &event); err == nil {
			h.impressions = append(h.impressions, event)
		}
	case "events.clicks":
		var event models.Click
		if err := json.Unmarshal(msg.Value, &event); err == nil {
			h.clicks = append(h.clicks, event)
		}
	case "events.conversions":
		var event models.Conversion
		if err := json.Unmarshal(msg.Value, &event); err == nil {
			h.conversions = append(h.conversions, event)
		}
	}
}

func (h *ConsumerHandler) flush(ctx context.Context) error {
	h.mu.Lock()
	imps := h.impressions
	clks := h.clicks
	convs := h.conversions
	h.impressions = nil
	h.clicks = nil
	h.conversions = nil
	h.lastFlush = time.Now()
	h.mu.Unlock()

	if len(imps) == 0 && len(clks) == 0 && len(convs) == 0 {
		return nil
	}

	log.Printf("Flushing batch: %d impressions, %d clicks, %d conversions", len(imps), len(clks), len(convs))
	return h.store.WriteBatch(ctx, imps, clks, convs)
}
