package store

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type AnalyticsStore struct{}

func NewAnalyticsStore() *AnalyticsStore {
	return &AnalyticsStore{}
}

func (s *AnalyticsStore) InsertClick(ctx *gofr.Context, click *model.ClickEvent) error {
	click.ID = primitive.NewObjectID().Hex()
	if click.Timestamp.IsZero() {
		click.Timestamp = time.Now().UTC()
	}
	_, err := ctx.Mongo.InsertOne(ctx, "click_events", click)
	return err
}

func (s *AnalyticsStore) HasIPClicked(ctx *gofr.Context, shortCode, ip string) (bool, error) {
	var result model.ClickEvent
	err := ctx.Mongo.FindOne(ctx, "click_events", bson.M{"short_code": shortCode, "ip_address": ip}, &result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
