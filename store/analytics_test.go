package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

func TestAnalyticsStoreGetFieldBreakdown(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	analyticsStore := store.NewAnalyticsStore()

	mocks.Mongo.EXPECT().Find(
		gomock.Any(),
		"click_events",
		bson.M{"short_code": "test-code"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		events, ok := res.(*[]model.ClickEvent)
		if ok {
			*events = []model.ClickEvent{
				{Browser: "Chrome"},
				{Browser: "Chrome"},
				{Browser: "Safari"},
			}
		}
		return nil
	})

	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	results, err := analyticsStore.GetFieldBreakdown(ctx, "test-code", "browser", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Chrome", results[0].Name)
	assert.Equal(t, int64(2), results[0].Count)
	assert.Equal(t, "Safari", results[1].Name)
	assert.Equal(t, int64(1), results[1].Count)
}

func TestAnalyticsStoreGetTimeseries(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	analyticsStore := store.NewAnalyticsStore()

	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

	mocks.Mongo.EXPECT().Find(
		gomock.Any(),
		"click_events",
		bson.M{"short_code": "test-code"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		events, ok := res.(*[]model.ClickEvent)
		if ok {
			*events = []model.ClickEvent{
				{Timestamp: now},
				{Timestamp: now},
			}
		}
		return nil
	})

	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	results, err := analyticsStore.GetTimeseries(ctx, "test-code", "day", 30)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "2026-08-26", results[0].Timestamp)
	assert.Equal(t, int64(2), results[0].Clicks)
}
