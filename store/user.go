package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

const usersCollection = "users"

type UserStore struct{}

func NewUserStore() *UserStore {
	return &UserStore{}
}

func (s *UserStore) Insert(ctx *gofr.Context, user *model.User) error {
	now := time.Now().UTC()
	user.ID = primitive.NewObjectID().Hex()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := ctx.Mongo.InsertOne(ctx, usersCollection, user)

	return err
}

func (s *UserStore) FindByEmailOrUsername(ctx *gofr.Context, identifier string) (*model.User, error) {
	var result model.User
	err := ctx.Mongo.FindOne(ctx, usersCollection, bson.M{
		"$or": []bson.M{
			{"email": identifier},
			{"username": identifier},
		},
	}, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *UserStore) FindByID(ctx *gofr.Context, id string) (*model.User, error) {
	var result model.User
	err := ctx.Mongo.FindOne(ctx, usersCollection, bson.M{"_id": id}, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *UserStore) UpdateProfile(ctx *gofr.Context, id string, update model.UserProfileUpdate) (*model.User, error) {
	set := bson.M{"updated_at": time.Now().UTC()}
	if update.Username != "" {
		set["username"] = update.Username
	}
	if update.Email != "" {
		set["email"] = update.Email
	}
	if update.PasswordHash != "" {
		set["password_hash"] = update.PasswordHash
	}

	if err := ctx.Mongo.UpdateOne(ctx, usersCollection, bson.M{"_id": id}, bson.M{"$set": set}); err != nil {
		return nil, err
	}

	return s.FindByID(ctx, id)
}

func (s *UserStore) AddAPIKey(ctx *gofr.Context, id, apiKey string) error {
	return ctx.Mongo.UpdateOne(ctx, usersCollection, bson.M{"_id": id}, bson.M{
		"$push": bson.M{"api_keys": apiKey},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	})
}

func (s *UserStore) RemoveAPIKey(ctx *gofr.Context, id, apiKey string) error {
	return ctx.Mongo.UpdateOne(ctx, usersCollection, bson.M{"_id": id}, bson.M{
		"$pull": bson.M{"api_keys": apiKey},
		"$set":  bson.M{"updated_at": time.Now().UTC()},
	})
}
