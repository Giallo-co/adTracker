package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"adtracker/internal/kafka"
	"adtracker/internal/models"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

func ImpressionHandler(c *gin.Context) {
	var event models.Impression
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ReceivedAt = time.Now()

	msg := &sarama.ProducerMessage{
		Topic: "events.impressions",
		Key:   sarama.StringEncoder(event.SessionID),
		Value: sarama.ByteEncoder(toJSON(event)),
	}

	select {
	case kafka.AsyncProducer.Input() <- msg:
		c.Status(http.StatusAccepted)
	default:
		// Backpressure handling: local fallback
		go kafka.SaveToFallback(msg)
		c.Status(http.StatusAccepted) // Still return 202 as per requirement
	}
}

func ClickHandler(c *gin.Context) {
	var event models.Click
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ReceivedAt = time.Now()

	msg := &sarama.ProducerMessage{
		Topic: "events.clicks",
		Key:   sarama.StringEncoder(event.UserInfo.SessionID),
		Value: sarama.ByteEncoder(toJSON(event)),
	}

	select {
	case kafka.AsyncProducer.Input() <- msg:
		c.Status(http.StatusAccepted)
	default:
		go kafka.SaveToFallback(msg)
		c.Status(http.StatusAccepted)
	}
}

func ConversionHandler(c *gin.Context) {
	var event models.Conversion
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ReceivedAt = time.Now()

	msg := &sarama.ProducerMessage{
		Topic: "events.conversions",
		Key:   sarama.StringEncoder(event.UserInfo.SessionID),
		Value: sarama.ByteEncoder(toJSON(event)),
	}

	select {
	case kafka.AsyncProducer.Input() <- msg:
		c.Status(http.StatusAccepted)
	default:
		go kafka.SaveToFallback(msg)
		c.Status(http.StatusAccepted)
	}
}

func toJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
