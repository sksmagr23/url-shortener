package model

import "time"

type URL struct {
	ID           string     `bson:"_id,omitempty"       json:"id"`
	Original     string     `bson:"original_url"        json:"original_url"`
	ShortCode    string     `bson:"short_code"          json:"short_code"`
	UserID       string     `bson:"user_id"             json:"user_id,omitempty"`
	Public       bool       `bson:"public"              json:"public"`
	CustomDomain string     `bson:"custom_domain,omitempty" json:"custom_domain,omitempty"`
	ExpiresAt    *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	TotalClicks  int64      `bson:"total_clicks"        json:"total_clicks"`
	UniqueClicks int64      `bson:"unique_clicks"       json:"unique_clicks"`
	CreatedAt    time.Time  `bson:"created_at"          json:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at"          json:"updated_at"`
	ShortURL     string     `bson:"-"                   json:"short_url"`
}

type URLUpdate struct {
	Original          string
	Public            *bool
	CustomDomain      *string
	ClearCustomDomain bool
	ExpiresAt         *time.Time
	ClearExpiry       bool
}

type URLListOptions struct {
	Page  int
	Limit int
	Sort  string
	Order string
}

type URLListResult struct {
	URLs       []*URL     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
