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
