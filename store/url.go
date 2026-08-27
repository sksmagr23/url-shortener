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

type URLStore struct {
	Mongo container.Mongo
}

func NewURLStore() *URLStore {
	return &URLStore{}
}

func NewURLStoreWithMongo(m container.Mongo) *URLStore {
	return &URLStore{Mongo: m}
}

func (s *URLStore) getMongo(ctx *gofr.Context) container.Mongo {
	if ctx != nil && ctx.Container != nil && ctx.Mongo != nil {
		return ctx.Mongo
	}
	if s != nil && s.Mongo != nil {
		return s.Mongo
	}
	return nil
}

func (s *URLStore) Insert(ctx *gofr.Context, url *model.URL) error {
	m := s.getMongo(ctx)
	if m == nil {
		return errors.New("mongodb datasource is not available")
	}
	now := time.Now().UTC()
	url.ID = primitive.NewObjectID().Hex()
	url.CreatedAt = now
	url.UpdatedAt = now
	_, err := m.InsertOne(ctx, "urls", url)
	return err
}

func (s *URLStore) FindByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result model.URL
	err := m.FindOne(ctx, "urls", bson.M{"short_code": code, "user_id": userID}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *URLStore) FindPublicByShortCode(ctx *gofr.Context, code string) (*model.URL, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result model.URL
	err := m.FindOne(ctx, "urls", bson.M{"short_code": code}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *URLStore) ListByUser(ctx *gofr.Context, userID string) ([]*model.URL, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result []*model.URL
	err := m.Find(ctx, "urls", bson.M{"user_id": userID}, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *URLStore) CountByUser(ctx *gofr.Context, userID string) (int64, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return 0, errors.New("mongodb datasource is not available")
	}
	return m.CountDocuments(ctx, "urls", bson.M{"user_id": userID})
}

func (s *URLStore) ListPublic(ctx *gofr.Context) ([]*model.URL, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
	var result []*model.URL
	err := m.Find(ctx, "urls", bson.M{"public": true}, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *URLStore) CountPublic(ctx *gofr.Context) (int64, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return 0, errors.New("mongodb datasource is not available")
	}
	return m.CountDocuments(ctx, "urls", bson.M{"public": true})
}

func (s *URLStore) UpdateByShortCode(ctx *gofr.Context, userID, code string, update model.URLUpdate) (*model.URL, error) {
	m := s.getMongo(ctx)
	if m == nil {
		return nil, errors.New("mongodb datasource is not available")
	}
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

	if err := m.UpdateOne(ctx, "urls", bson.M{"short_code": code, "user_id": userID}, mutation); err != nil {
		return nil, err
	}

	return s.FindByShortCode(ctx, userID, code)
}

func (s *URLStore) DeleteByShortCode(ctx *gofr.Context, userID, code string) error {
	m := s.getMongo(ctx)
	if m == nil {
		return errors.New("mongodb datasource is not available")
	}
	_, err := m.DeleteOne(ctx, "urls", bson.M{"short_code": code, "user_id": userID})

	return err
}

func (s *URLStore) IncrementClicks(ctx *gofr.Context, code string, isUnique bool) error {
	m := s.getMongo(ctx)
	if m == nil {
		return errors.New("mongodb datasource is not available")
	}
	inc := bson.M{"total_clicks": 1}
	if isUnique {
		inc["unique_clicks"] = 1
	}
	return m.UpdateOne(ctx, "urls", bson.M{"short_code": code}, bson.M{
		"$inc": inc,
		"$set": bson.M{"updated_at": time.Now().UTC()},
	})
}
