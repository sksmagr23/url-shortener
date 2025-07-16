package model

import "time"

type User struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Username     string    `bson:"username"      json:"username"`
	Email        string    `bson:"email"         json:"email"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	APIKey       string    `bson:"api_key"       json:"api_key,omitempty"`
	CreatedAt    time.Time `bson:"created_at"    json:"created_at"`
}
