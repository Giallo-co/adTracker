package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"adtracker/internal/models"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	influxClient influxdb2.Client
	writeAPI     api.WriteAPIBlocking
	minioClient  *minio.Client
	minioBucket  string
	org          string
	bucket       string
}

func NewStorage() *Storage {
	influxURL := os.Getenv("INFLUXDB_URL")
	token := os.Getenv("INFLUXDB_TOKEN")
	org := os.Getenv("INFLUXDB_ORG")
	bucket := os.Getenv("INFLUXDB_BUCKET")

	client := influxdb2.NewClient(influxURL, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")

	mClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Error creating MinIO client: %v", err)
	}

	return &Storage{
		influxClient: client,
		writeAPI:     writeAPI,
		minioClient:  mClient,
		minioBucket:  minioBucket,
		org:          org,
		bucket:       bucket,
	}
}

func (s *Storage) WriteBatch(ctx context.Context, impressions []models.Impression, clicks []models.Click, conversions []models.Conversion) error {
	// 1. Write to MinIO (Hive-style)
	if err := s.writeToMinIO(ctx, "impressions", impressions); err != nil {
		return fmt.Errorf("minio impressions error: %w", err)
	}
	if err := s.writeToMinIO(ctx, "clicks", clicks); err != nil {
		return fmt.Errorf("minio clicks error: %w", err)
	}
	if err := s.writeToMinIO(ctx, "conversions", conversions); err != nil {
		return fmt.Errorf("minio conversions error: %w", err)
	}

	// 2. Write to InfluxDB
	if err := s.writeToInflux(ctx, impressions, clicks, conversions); err != nil {
		return fmt.Errorf("influx error: %w", err)
	}

	return nil
}

func (s *Storage) writeToMinIO(ctx context.Context, eventType string, events interface{}) error {
	// For simplicity, we write one file per batch.
	// Hive-style: year=YYYY/month=MM/day=DD/hour=HH/
	now := time.Now()
	path := fmt.Sprintf("events/%s/year=%d/month=%02d/day=%02d/hour=%02d/%d.json",
		eventType, now.Year(), now.Month(), now.Day(), now.Hour(), now.UnixNano())

	data, err := models.ToJSON(events)
	if err != nil {
		return err
	}

	_, err = s.minioClient.PutObject(ctx, s.minioBucket, path, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	return err
}

func (s *Storage) writeToInflux(ctx context.Context, impressions []models.Impression, clicks []models.Click, conversions []models.Conversion) error {
	for _, imp := range impressions {
		p := influxdb2.NewPoint("impression",
			map[string]string{
				"state":           imp.State,
				"search_keywords": imp.SearchKeywords,
				"session_id":      imp.SessionID,
			},
			map[string]interface{}{
				"count": 1,
			},
			imp.Timestamp)
		if err := s.writeAPI.WritePoint(ctx, p); err != nil {
			return err
		}
	}

	for _, clk := range clicks {
		p := influxdb2.NewPoint("click",
			map[string]string{
				"state":      clk.UserInfo.State,
				"ad_id":      clk.ClickedAd.AdID,
				"session_id": clk.UserInfo.SessionID,
			},
			map[string]interface{}{
				"count":         1,
				"time_to_click": clk.ClickedAd.TimeToClick,
			},
			clk.Timestamp)
		if err := s.writeAPI.WritePoint(ctx, p); err != nil {
			return err
		}
	}

	for _, conv := range conversions {
		p := influxdb2.NewPoint("conversion",
			map[string]string{
				"state":           conv.UserInfo.State,
				"conversion_type": conv.ConversionType,
				"session_id":      conv.UserInfo.SessionID,
			},
			map[string]interface{}{
				"count":           1,
				"revenue":         conv.ConversionValue,
				"time_to_convert": conv.AttributionInfo.TimeToConvert,
			},
			conv.Timestamp)
		if err := s.writeAPI.WritePoint(ctx, p); err != nil {
			return err
		}
	}

	return nil
}
