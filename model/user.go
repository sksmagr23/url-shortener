package model

import "time"

type User struct {
	ID           string    `bson:"_id,omitempty"        json:"id"`
	Username     string    `bson:"username"             json:"username"`
	Email        string    `bson:"email"                json:"email"`
	PasswordHash string    `bson:"password_hash"        json:"-"`
	APIKeys      []string  `bson:"api_keys,omitempty"   json:"api_keys,omitempty"`
	CreatedAt    time.Time `bson:"created_at"           json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"           json:"updated_at"`
}

type UserProfileUpdate struct {
	Username     string
	Email        string
	PasswordHash string
}
