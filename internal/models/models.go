package models

import (
	"encoding/json"
	"time"
)

// Impression represents the event of an ad being viewed.
type Impression struct {
	ImpressionID   string    `json:"impression_id" binding:"required"`
	UserIP         string    `json:"user_ip" binding:"required,ip"`
	UserAgent      string    `json:"user_agent"`
	Timestamp      time.Time `json:"timestamp" binding:"required"` // RFC3339
	State          string    `json:"state" binding:"required"`      // Key for aggregation
	SearchKeywords string    `json:"search_keywords"`               // Key for aggregation
	SessionID      string    `json:"session_id" binding:"required"`
	Ads            []AdEntry `json:"ads" binding:"required,dive"`
	ReceivedAt     time.Time `json:"received_at"` // Internal metadata
}

type AdEntry struct {
	Advertiser Advertiser `json:"advertiser"`
	Campaign   Campaign   `json:"campaign"`
	Ad         AdDetail   `json:"ad"`
}

type Advertiser struct {
	AdvertiserID   string `json:"advertiser_id" binding:"required"`
	AdvertiserName string `json:"advertiser_name"`
}

type Campaign struct {
	CampaignID   string `json:"campaign_id" binding:"required"`
	CampaignName string `json:"campaign_name"`
}

type AdDetail struct {
	AdID       string `json:"ad_id" binding:"required"`
	AdName     string `json:"ad_name"`
	AdText     string `json:"ad_text"`
	AdLink     string `json:"ad_link"`
	AdPosition int    `json:"ad_position"`
	AdFormat   string `json:"ad_format"`
}

// Click represents the user interaction with an ad.
type Click struct {
	ClickID      string    `json:"click_id" binding:"required"`
	ImpressionID string    `json:"impression_id" binding:"required"`
	Timestamp    time.Time `json:"timestamp" binding:"required"`
	ClickedAd    struct {
		AdID             string  `json:"ad_id" binding:"required"`
		AdPosition       int     `json:"ad_position"`
		ClickCoordinates struct {
			X           int     `json:"x"`
			Y           int     `json:"y"`
			NormalizedX float64 `json:"normalized_x"`
			NormalizedY float64 `json:"normalized_y"`
		} `json:"click_coordinates"`
		TimeToClick float64 `json:"time_to_click"` // Used for avg metrics
	} `json:"clicked_ad" binding:"required"`
	UserInfo   UserInfo  `json:"user_info" binding:"required"`
	ReceivedAt time.Time `json:"received_at"` // Internal metadata
}

// Conversion represents the final valuable action of a user.
type Conversion struct {
	ConversionID       string    `json:"conversion_id" binding:"required"`
	ClickID            string    `json:"click_id" binding:"required"`
	ImpressionID       string    `json:"impression_id" binding:"required"`
	Timestamp          time.Time `json:"timestamp" binding:"required"`
	ConversionType     string    `json:"conversion_type"`
	ConversionValue    float64   `json:"conversion_value"` // For revenue metric
	ConversionCurrency string    `json:"conversion_currency"`
	ConversionAttributes struct {
		OrderID string `json:"order_id"`
		Items   []struct {
			ProductID string  `json:"product_id"`
			Quantity  int     `json:"quantity"`
			UnitPrice float64 `json:"unit_price"`
		} `json:"items"`
	} `json:"conversion_attributes"`
	AttributionInfo struct {
		TimeToConvert    float64 `json:"time_to_convert"` // Used for avg metrics
		AttributionModel string  `json:"attribution_model"`
	} `json:"attribution_info"`
	UserInfo   UserInfo  `json:"user_info" binding:"required"`
	ReceivedAt time.Time `json:"received_at"` // Internal metadata
}

type UserInfo struct {
	UserIP    string `json:"user_ip" binding:"required,ip"`
	State     string `json:"state" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
}

func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
