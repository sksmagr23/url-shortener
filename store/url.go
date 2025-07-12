package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type URLStore struct{}

func NewURLStore() *URLStore {
	return &URLStore{}
}

func (s *URLStore) Insert(ctx *gofr.Context, url *model.URL) error {
	url.CreatedAt = time.Now().UTC()
	_, err := ctx.Mongo.InsertOne(ctx, "urls", url)
	return err
}

func (s *URLStore) FindByShortCode(ctx *gofr.Context, code string) (*model.URL, error) {
	var result model.URL
	err := ctx.Mongo.FindOne(ctx, "urls", bson.M{"short_code": code}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
