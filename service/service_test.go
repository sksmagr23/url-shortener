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
	"github.com/redis/go-redis/v9"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"

	"github.com/sksmagr23/url-shortener-gofr/auth"
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
			analyticsStore := store.NewAnalyticsStore()
			urlService := service.NewURLService(urlStore, analyticsStore, nil, tt.host)

			if !tt.expectError {
				mocks.Mongo.EXPECT().FindOne(
					gomock.Any(),
					"urls",
					gomock.Any(),
					gomock.Any(),
				).Return(mongo.ErrNoDocuments)

				mocks.Mongo.EXPECT().InsertOne(
					gomock.Any(),
					"urls",
					gomock.Any(),
				).Return("test-id", nil)
			}

			ctx := &gofr.Context{
				Context:   auth.ContextWithUserID(context.Background(), "user-1"),
				Container: mockContainer,
			}

			result, err := urlService.Create(ctx, "user-1", service.URLCreateInput{Original: tt.originalURL})

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.originalURL, result.Original)
			assert.Equal(t, "user-1", result.UserID)
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
			analyticsStore := store.NewAnalyticsStore()
			urlService := service.NewURLService(urlStore, analyticsStore, nil, tt.host)

			if tt.mockError != nil {
				mocks.Mongo.EXPECT().FindOne(
					gomock.Any(),
					"urls",
					bson.M{"short_code": tt.shortCode, "user_id": "user-1"},
					gomock.Any(),
				).Return(tt.mockError)
			} else {
				mocks.Mongo.EXPECT().FindOne(
					gomock.Any(),
					"urls",
					bson.M{"short_code": tt.shortCode, "user_id": "user-1"},
					gomock.Any(),
				).Return(nil)
			}

			ctx := &gofr.Context{
				Context:   auth.ContextWithUserID(context.Background(), "user-1"),
				Container: mockContainer,
			}

			result, err := urlService.GetByShortCode(ctx, "user-1", tt.shortCode)
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
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
	).Return("", errors.New("database connection failed"))
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
		gomock.Any(),
	).Return(mongo.ErrNoDocuments)

	ctx := &gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}

	result, err := urlService.Create(ctx, "user-1", service.URLCreateInput{Original: "https://example.com/test"})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestURLServiceGetByShortCodeWithDatabaseError(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "test123", "user_id": "user-1"},
		gomock.Any(),
	).Return(errors.New("database connection failed"))

	ctx := &gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}

	result, err := urlService.GetByShortCode(ctx, "user-1", "test123")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestURLServiceCreateWithCustomCodeAndPublicFlag(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code"},
		gomock.Any(),
	).Return(mongo.ErrNoDocuments)
	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
	).Return("url-1", nil)

	result, err := urlService.Create(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}, "user-1", service.URLCreateInput{
		Original:   "https://example.com/test",
		CustomCode: "my-code",
		Public:     true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "my-code", result.ShortCode)
	assert.True(t, result.Public)
	assert.Equal(t, "http://localhost:8000/my-code", result.ShortURL)
}

func TestURLServiceCreateRejectsDuplicateCustomCode(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code"},
		gomock.Any(),
	).Return(nil)

	result, err := urlService.Create(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}, "user-1", service.URLCreateInput{
		Original:   "https://example.com/test",
		CustomCode: "my-code",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestURLServicePrivateRedirectRequiresOwner(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "private-code"},
		gomock.Any(),
	).Return(nil)

	result, err := urlService.GetRedirectByShortCode(&gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}, "other-user", "private-code")

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestURLServiceCreateWithCustomDomain(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code"},
		gomock.Any(),
	).Return(mongo.ErrNoDocuments)
	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
	).Return("url-1", nil)

	result, err := urlService.Create(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}, "user-1", service.URLCreateInput{
		Original:     "https://example.com/test",
		CustomCode:   "my-code",
		CustomDomain: "my.tech",
	})

	assert.NoError(t, err)
	assert.Equal(t, "my-code", result.ShortCode)
	assert.Equal(t, "my.tech", result.CustomDomain)
	assert.Equal(t, "https://my.tech/my-code", result.ShortURL)
}

func TestURLServiceUpdateCustomDomain(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	customDomain := "custom.example.com"

	mocks.Mongo.EXPECT().UpdateOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code", "user_id": "user-1"},
		gomock.Any(),
	).Return(nil)

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code", "user_id": "user-1"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		u, ok := res.(*model.URL)
		if ok {
			u.ShortCode = "my-code"
			u.CustomDomain = "custom.example.com"
		}
		return nil
	})

	result, err := urlService.Update(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}, "user-1", "my-code", service.URLUpdateInput{
		CustomDomain: &customDomain,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "custom.example.com", result.CustomDomain)
	assert.Equal(t, "https://custom.example.com/my-code", result.ShortURL)
}

func TestURLServiceGetRedirectRecordsClick(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	// Expect FindPublicByShortCode query
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		u, ok := res.(*model.URL)
		if ok {
			u.ID = "url-123"
			u.ShortCode = "my-code"
			u.Public = true
		}
		return nil
	})

	// Expect HasIPClicked query
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"click_events",
		bson.M{"short_code": "my-code", "ip_address": "203.0.113.195"},
		gomock.Any(),
	).Return(mongo.ErrNoDocuments)

	// Expect InsertClick write
	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		"click_events",
		gomock.Any(),
	).Return("click-123", nil)

	// Expect IncrementClicks write
	mocks.Mongo.EXPECT().UpdateOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code"},
		gomock.Any(),
	).Return(nil)

	// Set up context with metadata
	meta := auth.RequestMetadata{
		IPAddress:  "203.0.113.195",
		UserAgent:  "Mozilla",
		Referrer:   "https://google.com",
		Browser:    "Chrome",
		OS:         "Windows",
		DeviceType: "Desktop",
		Country:    "US",
	}
	ctx := &gofr.Context{
		Context:   auth.ContextWithMetadata(context.Background(), meta),
		Container: mockContainer,
	}

	result, err := urlService.GetRedirectByShortCode(ctx, "", "my-code")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "url-123", result.ID)
	time.Sleep(30 * time.Millisecond)
}

func TestURLServiceCacheHit(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlCache := store.NewURLCache()
	urlService := service.NewURLService(nil, nil, urlCache, "http://localhost:8000/")

	cachedURLJSON := `{"id":"cached-id","original_url":"https://cached.com","short_code":"cached-code","public":true}`

	mocks.Redis.EXPECT().Get(gomock.Any(), "url:cached-code").Return(redis.NewStringResult(cachedURLJSON, nil))

	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	result, err := urlService.GetRedirectByShortCode(ctx, "", "cached-code")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://cached.com", result.Original)
	assert.Equal(t, "http://localhost:8000/cached-code", result.ShortURL)
}

func TestURLServiceUpdateInvalidatesCache(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlCache := store.NewURLCache()
	urlService := service.NewURLService(urlStore, analyticsStore, urlCache, "http://localhost:8000/")

	customDomain := "custom.example.com"

	mocks.Mongo.EXPECT().UpdateOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code", "user_id": "user-1"},
		gomock.Any(),
	).Return(nil)

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "my-code", "user_id": "user-1"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		u, ok := res.(*model.URL)
		if ok {
			u.ShortCode = "my-code"
			u.CustomDomain = "custom.example.com"
		}
		return nil
	})

	mocks.Redis.EXPECT().Del(gomock.Any(), "url:my-code").Return(redis.NewIntResult(1, nil))

	result, err := urlService.Update(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}, "user-1", "my-code", service.URLUpdateInput{
		CustomDomain: &customDomain,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestURLServiceGetAnalyticsSummary(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	// Expect FindByShortCodeAndUserID
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "summary-code", "user_id": "user-1"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		u, ok := res.(*model.URL)
		if ok {
			u.ShortCode = "summary-code"
			u.UserID = "user-1"
			u.TotalClicks = 25
			u.UniqueClicks = 10
		}
		return nil
	})

	// Expect 5 breakdown Find calls (browser, os, device_type, country, referrer)
	mocks.Mongo.EXPECT().Find(
		gomock.Any(),
		"click_events",
		bson.M{"short_code": "summary-code"},
		gomock.Any(),
	).Times(5).Return(nil)

	ctx := &gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}

	res, err := urlService.GetAnalyticsSummary(ctx, "user-1", "summary-code")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "summary-code", res.ShortCode)
	assert.Equal(t, int64(25), res.TotalClicks)
	assert.Equal(t, int64(10), res.UniqueClicks)
}

func TestURLServiceGetAnalyticsTimeseries(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	// Expect FindByShortCodeAndUserID
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "ts-code", "user_id": "user-1"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		u, ok := res.(*model.URL)
		if ok {
			u.ShortCode = "ts-code"
			u.UserID = "user-1"
		}
		return nil
	})

	// Expect 1 timeseries Find call
	mocks.Mongo.EXPECT().Find(
		gomock.Any(),
		"click_events",
		bson.M{"short_code": "ts-code"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		events, ok := res.(*[]model.ClickEvent)
		if ok {
			*events = []model.ClickEvent{{Timestamp: time.Now().UTC()}}
		}
		return nil
	})

	ctx := &gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}

	res, err := urlService.GetAnalyticsTimeseries(ctx, "user-1", "ts-code", "day", 30)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "ts-code", res.ShortCode)
	assert.Equal(t, "day", res.Unit)
	assert.Len(t, res.Timeseries, 1)
}

func TestURLServiceGetQRCode(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)
	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "qr-code"},
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		u, ok := res.(*model.URL)
		if ok {
			u.ShortCode = "qr-code"
			u.Public = true
			u.Original = "https://gofr.dev"
		}
		return nil
	})

	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	qrResp, err := urlService.GetQRCode(ctx, "", "qr-code", 256)
	assert.NoError(t, err)
	assert.NotNil(t, qrResp)
	assert.Equal(t, "qr-code", qrResp.ShortCode)
	assert.True(t, len(qrResp.PNGBytes) > 0)
	assert.True(t, len(qrResp.QRCodeBase64) > 0)
}
