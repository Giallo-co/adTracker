package main

import (
	"log"
	"os"
	"strings"

	"adtracker/internal/handlers"
	"adtracker/internal/kafka"

	"github.com/gin-gonic/gin"
)

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	brokerList := strings.Split(brokers, ",")

	kafka.InitProducer(brokerList)
	defer kafka.AsyncProducer.Close()

	r := gin.Default()

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	api := r.Group("/api/events")
	{
		api.POST("/impression", handlers.ImpressionHandler)
		api.POST("/click", handlers.ClickHandler)
		api.POST("/conversion", handlers.ConversionHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
