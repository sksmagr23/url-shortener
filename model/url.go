package model

import "time"

type URL struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Original  string    `bson:"original_url"  json:"original_url"`
	ShortCode string    `bson:"short_code"    json:"short_code"`
	CreatedAt time.Time `bson:"created_at"    json:"created_at"`
	ShortURL  string    `bson:"-"             json:"short_url"`
}
