package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/service"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

func TestGenerateShortCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"Default Length", 6},
		{"Custom Length 8", 8},
		{"Custom Length 10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := service.GenerateShortCode(tt.length)
			assert.Equal(t, tt.length, len(code))

			validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, char := range code {
				assert.True(t, strings.ContainsRune(validChars, char),
					"Character '%c' is not in valid charset", char)
			}
		})
	}
}

func TestURLServiceCreate(t *testing.T) {
	tests := []struct {
		name          string
		originalURL   string
		host          string
		expectError   bool
		expectedError string
	}{
		{
			name:        "Valid HTTPS URL",
			originalURL: "https://example.com/test",
			host:        "http://localhost:8000/",
			expectError: false,
		},
		{
			name:        "Valid HTTP URL",
			originalURL: "http://example.com/test",
			host:        "http://localhost:8000/",
			expectError: false,
		},
		{
			name:          "Invalid URL - No Protocol",
			originalURL:   "example.com/test",
			host:          "http://localhost:8000/",
			expectError:   true,
			expectedError: "invalid URL",
		},
		{
			name:          "Invalid URL - Empty",
			originalURL:   "",
			host:          "http://localhost:8000/",
			expectError:   true,
			expectedError: "invalid URL",
		},
		{
			name:        "Custom Host",
			originalURL: "https://example.com/test",
			host:        "https://myshortener.com/",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockContainer, mocks := container.NewMockContainer(t)
			urlStore := store.NewURLStore()
			urlService := service.NewURLService(urlStore, tt.host)

			if !tt.expectError {
				mocks.Mongo.EXPECT().InsertOne(
					gomock.Any(),
					"urls",
					gomock.Any(),
				).Return("test-id", nil)
			}

			ctx := &gofr.Context{
				Context:   context.Background(),
				Container: mockContainer,
			}

			result, err := urlService.Create(ctx, tt.originalURL)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.originalURL, result.Original)
			assert.NotEmpty(t, result.ShortCode)
			assert.Len(t, result.ShortCode, 6)
			assert.NotEmpty(t, result.ShortURL)
			assert.NotZero(t, result.CreatedAt)

			assert.True(t, strings.HasPrefix(result.ShortURL, tt.host))
			assert.True(t, strings.HasSuffix(result.ShortURL, result.ShortCode))
		})
	}
}

func TestURLServiceGetByShortCode(t *testing.T) {
	tests := []struct {
		name        string
		shortCode   string
		mockURL     *model.URL
		mockError   error
		host        string
		expectError bool
	}{
		{
			name:      "Success - Valid Short Code",
			shortCode: "abc123",
			mockURL: &model.URL{
				ID:        "test-id",
				Original:  "https://example.com/test",
				ShortCode: "abc123",
				CreatedAt: time.Now().UTC(),
			},
			mockError:   nil,
			host:        "http://localhost:8000/",
			expectError: false,
		},
		{
			name:        "Failure - URL Not Found",
			shortCode:   "nonexistent",
			mockURL:     nil,
			mockError:   mongo.ErrNoDocuments,
			host:        "http://localhost:8000/",
			expectError: true,
		},
		{
			name:      "Custom Host",
			shortCode: "abc123",
			mockURL: &model.URL{
				ID:        "test-id",
				Original:  "https://example.com/test",
				ShortCode: "abc123",
				CreatedAt: time.Now().UTC(),
			},
			mockError:   nil,
			host:        "https://myshortener.com/",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockContainer, mocks := container.NewMockContainer(t)
			urlStore := store.NewURLStore()
			urlService := service.NewURLService(urlStore, tt.host)

			if tt.mockError != nil {
				mocks.Mongo.EXPECT().FindOne(
					gomock.Any(),
					"urls",
					bson.M{"short_code": tt.shortCode},
					gomock.Any(),
				).Return(tt.mockError)
			} else {
				mocks.Mongo.EXPECT().FindOne(
					gomock.Any(),
					"urls",
					bson.M{"short_code": tt.shortCode},
					gomock.Any(),
				).Return(nil)
			}

			ctx := &gofr.Context{
				Context:   context.Background(),
				Container: mockContainer,
			}

			result, err := urlService.GetByShortCode(ctx, tt.shortCode)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)

			result.Original = tt.mockURL.Original
			result.ShortCode = tt.mockURL.ShortCode
			result.ShortURL = tt.host + tt.mockURL.ShortCode

			assert.Equal(t, tt.mockURL.Original, result.Original)
			assert.Equal(t, tt.mockURL.ShortCode, result.ShortCode)
			assert.NotEmpty(t, result.ShortURL)
			assert.True(t, strings.HasPrefix(result.ShortURL, tt.host))
			assert.True(t, strings.HasSuffix(result.ShortURL, result.ShortCode))
		})
	}
}

func TestURLServiceCreateWithDatabaseError(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	urlService := service.NewURLService(urlStore, "http://localhost:8000/")

	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
	).Return("", errors.New("database connection failed"))

	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	result, err := urlService.Create(ctx, "https://example.com/test")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestURLServiceGetByShortCodeWithDatabaseError(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	urlService := service.NewURLService(urlStore, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "test123"},
		gomock.Any(),
	).Return(errors.New("database connection failed"))

	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	result, err := urlService.GetByShortCode(ctx, "test123")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection failed")
}
