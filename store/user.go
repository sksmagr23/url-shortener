package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type UserStore struct{}

func NewUserStore() *UserStore {
	return &UserStore{}
}

func (s *UserStore) Insert(ctx *gofr.Context, user *model.User) error {
	user.CreatedAt = time.Now().UTC()
	_, err := ctx.Mongo.InsertOne(ctx, "users", user)
	return err
}

func (s *UserStore) FindByEmail(ctx *gofr.Context, email string) (*model.User, error) {
	var user model.User
	err := ctx.Mongo.FindOne(ctx, "users", bson.M{"email": email}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) FindByUsername(ctx *gofr.Context, username string) (*model.User, error) {
	var user model.User
	err := ctx.Mongo.FindOne(ctx, "users", bson.M{"username": username}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) FindByID(ctx *gofr.Context, id string) (*model.User, error) {
	var user model.User
	err := ctx.Mongo.FindOne(ctx, "users", bson.M{"_id": id}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) Update(ctx *gofr.Context, id string, update bson.M) error {
	filter := bson.M{"_id": id}
	return ctx.Mongo.UpdateOne(ctx, "users", filter, bson.M{"$set": update})
}
