package store

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

const usersCollection = "users"

type UserStore struct {
	Mongo container.Mongo
}

func NewUserStore() *UserStore {
	return &UserStore{}
}

func NewUserStoreWithMongo(m container.Mongo) *UserStore {
	return &UserStore{Mongo: m}
}

func (s *UserStore) getMongo(ctx *gofr.Context) container.Mongo {
	if ctx != nil && ctx.Container != nil && ctx.Mongo != nil {
		return ctx.Mongo
	}
	if s != nil && s.Mongo != nil {
		return s.Mongo
	}
	return nil
}

func (s *UserStore) Insert(ctx *gofr.Context, user *model.User) error {
	m := s.getMongo(ctx)
	if m == nil {
		return errors.New("mongodb datasource is not available")
	}
	now := time.Now().UTC()
	user.ID = primitive.NewObjectID().Hex()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := m.InsertOne(ctx, usersCollection, user)

	return err
}

func (s *UserStore) FindByEmailOrUsername(ctx *gofr.Context, identifier string) (*model.User, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result model.User
	err := m.FindOne(ctx, usersCollection, bson.M{
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
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result model.User
	err := m.FindOne(ctx, usersCollection, bson.M{"_id": id}, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *UserStore) FindByAPIKey(ctx *gofr.Context, apiKey string) (*model.User, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result model.User
	err := m.FindOne(ctx, usersCollection, bson.M{"api_keys": apiKey}, &result)
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
