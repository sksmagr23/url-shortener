package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"

	gofrHttp "gofr.dev/pkg/gofr/http"

	"github.com/sksmagr23/url-shortener-gofr/handler"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockMongoError error
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "Success - MongoDB Connected",
			mockMongoError: nil,
			expectedStatus: "connected",
			expectError:    false,
		},
		{
			name:           "Failure - MongoDB Disconnected",
			mockMongoError: mongo.ErrNoDocuments,
			expectedStatus: "disconnected",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockContainer, mocks := container.NewMockContainer(t)

			if tt.mockMongoError != nil {
				mocks.Mongo.EXPECT().CountDocuments(
					gomock.Any(),
					"urls",
					map[string]interface{}{},
				).Return(int64(0), tt.mockMongoError)
			} else {
				mocks.Mongo.EXPECT().CountDocuments(
					gomock.Any(),
					"urls",
					map[string]interface{}{},
				).Return(int64(5), nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			request := gofrHttp.NewRequest(req)

			ctx := &gofr.Context{
				Context:   context.Background(),
				Request:   request,
				Container: mockContainer,
			}

			healthHandler := handler.HealthHandler()
			result, err := healthHandler(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			response, ok := result.(map[string]interface{})
			assert.True(t, ok, "Expected result to be map[string]interface{}")
			assert.Equal(t, "healthy", response["status"])
			assert.NotEmpty(t, response["timestamp"])
			services, ok := response["services"].(map[string]string)
			assert.True(t, ok, "Expected services to be map[string]string")
			assert.Equal(t, tt.expectedStatus, services["mongoDB"])
		})
	}
}

func TestHealthHandlerWithInvalidRequest(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)

	mocks.Mongo.EXPECT().CountDocuments(
		gomock.Any(),
		"urls",
		map[string]interface{}{},
	).Return(int64(0), mongo.ErrNoDocuments)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	request := gofrHttp.NewRequest(req)

	ctx := &gofr.Context{
		Context:   context.Background(),
		Request:   request,
		Container: mockContainer,
	}

	healthHandler := handler.HealthHandler()
	result, err := healthHandler(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	response, ok := result.(map[string]interface{})
	assert.True(t, ok)
	services, ok := response["services"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "disconnected", services["mongoDB"])
}
