package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Impression struct {
	ImpressionID   string    `json:"impression_id"`
	UserIP         string    `json:"user_ip"`
	UserAgent      string    `json:"user_agent"`
	Timestamp      time.Time `json:"timestamp"`
	State          string    `json:"state"`
	SearchKeywords string    `json:"search_keywords"`
	SessionID      string    `json:"session_id"`
	Ads            []AdEntry `json:"ads"`
}

type AdEntry struct {
	Advertiser Advertiser `json:"advertiser"`
	Campaign   Campaign   `json:"campaign"`
	Ad         AdDetail   `json:"ad"`
}

type Advertiser struct {
	AdvertiserID   string `json:"advertiser_id"`
	AdvertiserName string `json:"advertiser_name"`
}

type Campaign struct {
	CampaignID   string `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`
}

type AdDetail struct {
	AdID       string `json:"ad_id"`
	AdName     string `json:"ad_name"`
	AdText     string `json:"ad_text"`
	AdLink     string `json:"ad_link"`
	AdPosition int    `json:"ad_position"`
	AdFormat   string `json:"ad_format"`
}

type Click struct {
	ClickID      string    `json:"click_id"`
	ImpressionID string    `json:"impression_id"`
	Timestamp    time.Time `json:"timestamp"`
	ClickedAd    struct {
		AdID     string `json:"ad_id"`
		Position int    `json:"ad_position"`
	} `json:"clicked_ad"`
	UserInfo UserInfo `json:"user_info"`
}

type Conversion struct {
	ConversionID string    `json:"conversion_id"`
	ClickID      string    `json:"click_id"`
	ImpressionID string    `json:"impression_id"`
	Timestamp    time.Time `json:"timestamp"`
	Value        float64   `json:"conversion_value"`
	UserInfo     UserInfo  `json:"user_info"`
}

type UserInfo struct {
	UserIP    string `json:"user_ip"`
	State     string `json:"state"`
	SessionID string `json:"session_id"`
}

var (
	successCount int64
	failCount    int64
	states       = []string{"CA", "NY", "TX", "FL", "IL", "WA", "MA", "GA"}
	keywords     = []string{"shoes", "laptop", "phone", "coffee", "travel"}
)

func main() {
	rps := flag.Int("rps", 100, "Requests per second")
	duration := flag.Duration("duration", 10*time.Second, "Test duration")
	baseURL := flag.String("url", "http://localhost:8080/api/events", "Base API URL")
	flag.Parse()

	targetRPS := *rps
	fmt.Printf("Starting stress test: %d RPS for %v\n", targetRPS, *duration)

	start := time.Now()
	stop := time.After(*duration)
	ticker := time.NewTicker(time.Second / time.Duration(targetRPS))
	defer ticker.Stop()

	var wg sync.WaitGroup
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        2000,
			MaxIdleConnsPerHost: 2000,
		},
	}

loop:
	for {
		select {
		case <-stop:
			break loop
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				runIteration(client, *baseURL)
			}()
		}
	}

	wg.Wait()
	totalTime := time.Since(start)
	fmt.Printf("\nTest Completed in %v\n", totalTime)
	fmt.Printf("Total Success: %d\n", atomic.LoadInt64(&successCount))
	fmt.Printf("Total Fail:    %d\n", atomic.LoadInt64(&failCount))
	fmt.Printf("Actual RPS:    %.2f\n", float64(atomic.LoadInt64(&successCount)+atomic.LoadInt64(&failCount))/totalTime.Seconds())
}

func runIteration(client *http.Client, baseURL string) {
	sessionID := uuid.New().String()
	ip := fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1)
	state := states[rand.Intn(len(states))]
	now := time.Now()

	// 1. Impression
	impID := uuid.New().String()
	imp := Impression{
		ImpressionID:   impID,
		UserIP:         ip,
		Timestamp:      now,
		State:          state,
		SearchKeywords: keywords[rand.Intn(len(keywords))],
		SessionID:      sessionID,
		Ads: []AdEntry{
			{
				Advertiser: Advertiser{AdvertiserID: "adv-789", AdvertiserName: "Nike Inc."},
				Campaign:   Campaign{CampaignID: "camp-456", CampaignName: "Fall 2026"},
				Ad: AdDetail{
					AdID:       "ad-123",
					AdName:     "Air Max Pro",
					AdPosition: 1,
				},
			},
		},
	}
	
	if send(client, baseURL+"/impression", imp) {
		atomic.AddInt64(&successCount, 1)
	} else {
		atomic.AddInt64(&failCount, 1)
	}

	// 2. Click (chance)
	if rand.Float64() < 0.3 {
		clickID := uuid.New().String()
		click := Click{
			ClickID:      clickID,
			ImpressionID: impID,
			Timestamp:    now.Add(5 * time.Second),
		}
		click.ClickedAd.AdID = "ad-123"
		click.ClickedAd.Position = 1
		click.UserInfo.SessionID = sessionID
		click.UserInfo.State = state
		click.UserInfo.UserIP = ip

		if send(client, baseURL+"/click", click) {
			atomic.AddInt64(&successCount, 1)
		} else {
			atomic.AddInt64(&failCount, 1)
		}

		// 3. Conversion (chance)
		if rand.Float64() < 0.2 {
			conv := Conversion{
				ConversionID: uuid.New().String(),
				ClickID:      clickID,
				ImpressionID: impID,
				Timestamp:    now.Add(15 * time.Minute),
				Value:        59.99,
			}
			conv.UserInfo.SessionID = sessionID
			conv.UserInfo.State = state
			conv.UserInfo.UserIP = ip

			if send(client, baseURL+"/conversion", conv) {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}
	}
}

func send(client *http.Client, url string, data interface{}) bool {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusAccepted
}
