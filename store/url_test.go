package store_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

func TestURLStoreListByUser(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()

	mocks.Mongo.EXPECT().Find(
		gomock.Any(),
		"urls",
		bson.M{"user_id": "user-1"},
		gomock.Any(),
	).Return(nil)

	urls, err := urlStore.ListByUser(&gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}, "user-1")

	assert.NoError(t, err)
	assert.Nil(t, urls)
}

func TestURLStoreUpdateByShortCode(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	isPublic := false

	mocks.Mongo.EXPECT().UpdateOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "abc123", "user_id": "user-1"},
		gomock.Any(),
	).Return(nil)
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "abc123", "user_id": "user-1"},
		gomock.Any(),
	).Return(nil)

	url, err := urlStore.UpdateByShortCode(&gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}, "user-1", "abc123", model.URLUpdate{Public: &isPublic})

	assert.NoError(t, err)
	assert.NotNil(t, url)
}

func TestURLStoreDeleteByShortCode(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()

	mocks.Mongo.EXPECT().DeleteOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "abc123", "user_id": "user-1"},
	).Return(int64(1), nil)

	err := urlStore.DeleteByShortCode(&gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}, "user-1", "abc123")

	assert.NoError(t, err)
}
