package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type URLStore struct{}

func NewURLStore() *URLStore {
	return &URLStore{}
}

func (s *URLStore) Insert(ctx *gofr.Context, url *model.URL) error {
	now := time.Now().UTC()
	url.ID = primitive.NewObjectID().Hex()
	url.CreatedAt = now
	url.UpdatedAt = now
	_, err := ctx.Mongo.InsertOne(ctx, "urls", url)
	return err
}

func (s *URLStore) FindByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	var result model.URL
	err := ctx.Mongo.FindOne(ctx, "urls", bson.M{"short_code": code, "user_id": userID}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *URLStore) FindPublicByShortCode(ctx *gofr.Context, code string) (*model.URL, error) {
	var result model.URL
	err := ctx.Mongo.FindOne(ctx, "urls", bson.M{"short_code": code}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *URLStore) ListByUser(ctx *gofr.Context, userID string) ([]*model.URL, error) {
	var result []*model.URL
	err := ctx.Mongo.Find(ctx, "urls", bson.M{"user_id": userID}, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *URLStore) CountByUser(ctx *gofr.Context, userID string) (int64, error) {
	return ctx.Mongo.CountDocuments(ctx, "urls", bson.M{"user_id": userID})
}

func (s *URLStore) UpdateByShortCode(ctx *gofr.Context, userID, code string, update model.URLUpdate) (*model.URL, error) {
	set := bson.M{"updated_at": time.Now().UTC()}
	if update.Original != "" {
		set["original_url"] = update.Original
	}
	if update.Public != nil {
		set["public"] = *update.Public
	}
	if update.CustomDomain != nil {
		set["custom_domain"] = *update.CustomDomain
	}
	if update.ExpiresAt != nil {
		set["expires_at"] = update.ExpiresAt
	}

	mutation := bson.M{"$set": set}
	unset := bson.M{}
	if update.ClearExpiry {
		unset["expires_at"] = ""
	}
	if update.ClearCustomDomain {
		unset["custom_domain"] = ""
	}
	if len(unset) > 0 {
		mutation["$unset"] = unset
	}

	if err := ctx.Mongo.UpdateOne(ctx, "urls", bson.M{"short_code": code, "user_id": userID}, mutation); err != nil {
		return nil, err
	}

	return s.FindByShortCode(ctx, userID, code)
}

func (s *URLStore) DeleteByShortCode(ctx *gofr.Context, userID, code string) error {
	_, err := ctx.Mongo.DeleteOne(ctx, "urls", bson.M{"short_code": code, "user_id": userID})

	return err
}

func (s *URLStore) IncrementClicks(ctx *gofr.Context, code string) error {
	return ctx.Mongo.UpdateOne(ctx, "urls", bson.M{"short_code": code}, bson.M{
		"$inc": bson.M{"total_clicks": 1},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	})
}
