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

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// --- CONFIGURACIÓN DE CORS ---
	// Esto es vital para que React pueda comunicarse con Go
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// GRUPO DE EVENTOS (POST)
	api := r.Group("/api/events")
	{
		api.POST("/impression", handlers.ImpressionHandler)
		api.POST("/click", handlers.ClickHandler)
		api.POST("/conversion", handlers.ConversionHandler)
	}

	// NUEVA RUTA PARA EL DASHBOARD (GET)
	r.GET("/api/stats", handlers.GetStatsHandler)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}