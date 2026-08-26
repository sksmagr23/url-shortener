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

func TestUserStoreAddAndRemoveAPIKey(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	userStore := store.NewUserStore()
	ctx := &gofr.Context{Context: context.Background(), Container: mockContainer}

	mocks.Mongo.EXPECT().UpdateOne(
		gomock.Any(),
		"users",
		bson.M{"_id": "user-1"},
		gomock.Any(),
	).Return(nil)

	err := userStore.AddAPIKey(ctx, "user-1", "usk_test")
	assert.NoError(t, err)

	mocks.Mongo.EXPECT().UpdateOne(
		gomock.Any(),
		"users",
		bson.M{"_id": "user-1"},
		gomock.Any(),
	).Return(nil)

	err = userStore.RemoveAPIKey(ctx, "user-1", "usk_test")
	assert.NoError(t, err)
}

func TestUserStoreInsertSetsTimestamps(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	userStore := store.NewUserStore()

	mocks.Mongo.EXPECT().InsertOne(gomock.Any(), "users", gomock.Any()).Return("user-1", nil)

	user := &model.User{Username: "saksham", Email: "saksham@example.com", PasswordHash: "hash"}
	err := userStore.Insert(&gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}, user)

	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)
}
