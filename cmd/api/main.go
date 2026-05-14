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
	// Configuración de Kafka
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	brokerList := strings.Split(brokers, ",")

	// Inicialización del productor de Kafka
	kafka.InitProducer(brokerList)
	defer kafka.AsyncProducer.Close()

	// Configuración del modo de Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// --- CONFIGURACIÓN DE CORS ---
	// Vital para que el Dashboard en http://localhost:5173 pueda pedir datos
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

	// GRUPO DE EVENTOS (POST) - Usado por el Stress Test
	events := r.Group("/api/events")
	{
		events.POST("/impression", handlers.ImpressionHandler)
		events.POST("/click", handlers.ClickHandler)
		events.POST("/conversion", handlers.ConversionHandler)
	}

	// RUTA PARA EL DASHBOARD (GET) - Usado por App.jsx
	r.GET("/api/stats", handlers.GetStatsHandler)

	// Ruta de salud para Docker
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// --- CONFIGURACIÓN DEL PUERTO ---
	// IMPORTANTE: Dentro del contenedor Docker debe ser 8080 
	// para que el mapeo "8085:8080" de tu docker-compose funcione
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" 
	}

	log.Printf("API Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}