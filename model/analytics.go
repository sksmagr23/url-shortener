package model

import "time"

type ClickEvent struct {
	ID         string    `bson:"_id,omitempty" json:"id"`
	URLID      string    `bson:"url_id"        json:"url_id"`
	ShortCode  string    `bson:"short_code"    json:"short_code"`
	Timestamp  time.Time `bson:"timestamp"     json:"timestamp"`
	IPAddress  string    `bson:"ip_address"    json:"ip_address"`
	UserAgent  string    `bson:"user_agent"    json:"user_agent"`
	Browser    string    `bson:"browser"       json:"browser"`
	OS         string    `bson:"os"            json:"os"`
	DeviceType string    `bson:"device_type"   json:"device_type"`
	Country    string    `bson:"country"       json:"country"`
	Referrer   string    `bson:"referrer"      json:"referrer"`
}

type BreakdownItem struct {
	Name  string `json:"name"  bson:"_id"`
	Count int64  `json:"count" bson:"count"`
}

type AnalyticsSummaryResponse struct {
	ShortCode    string          `json:"short_code"`
	TotalClicks  int64           `json:"total_clicks"`
	UniqueClicks int64           `json:"unique_clicks"`
	Browsers     []BreakdownItem `json:"browsers"`
	OS           []BreakdownItem `json:"os"`
	Devices      []BreakdownItem `json:"devices"`
	Countries    []BreakdownItem `json:"countries"`
	Referrers    []BreakdownItem `json:"referrers"`
}

type TimeseriesPoint struct {
	Timestamp string `json:"timestamp" bson:"_id"`
	Clicks    int64  `json:"clicks"    bson:"count"`
}

type AnalyticsTimeseriesResponse struct {
	ShortCode  string            `json:"short_code"`
	Unit       string            `json:"unit"`
	Timeseries []TimeseriesPoint `json:"timeseries"`
}
