package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"adtracker/internal/kafka"
	"adtracker/internal/models"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

// Variables globales para contadores en memoria
var (
	TotalImpressions uint64
	TotalClicks      uint64
	TotalConversions uint64
	StateCounts      = make(map[string]int)
)

// GetStatsHandler envía los datos al Dashboard de React
func GetStatsHandler(c *gin.Context) {
	stateReport := []gin.H{}
	
	for state, count := range StateCounts {
		stateReport = append(stateReport, gin.H{
			"name":  state,
			"value": count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"impressions": atomic.LoadUint64(&TotalImpressions),
		"clicks":      atomic.LoadUint64(&TotalClicks),
		"conversions": atomic.LoadUint64(&TotalConversions),
		"stateReport": stateReport,
	})
}

func ImpressionHandler(c *gin.Context) {
	var event models.Impression
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Incremento atómico del contador global
	atomic.AddUint64(&TotalImpressions, 1)
	
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
		go kafka.SaveToFallback(msg)
		c.Status(http.StatusAccepted)
	}
}

func ClickHandler(c *gin.Context) {
	var event models.Click
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	atomic.AddUint64(&TotalClicks, 1)

	event.ReceivedAt = time.Now()

	// Nota: He simplificado la Key para evitar errores con UserInfo
	msg := &sarama.ProducerMessage{
		Topic: "events.clicks",
		Key:   sarama.StringEncoder(time.Now().String()), 
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

	atomic.AddUint64(&TotalConversions, 1)

	event.ReceivedAt = time.Now()

	msg := &sarama.ProducerMessage{
		Topic: "events.conversions",
		Key:   sarama.StringEncoder(time.Now().String()),
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