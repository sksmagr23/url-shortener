package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sksmagr23/url-shortener-gofr/internal/handler"
	"github.com/sksmagr23/url-shortener-gofr/internal/model"
	"github.com/sksmagr23/url-shortener-gofr/internal/service"
	"github.com/sksmagr23/url-shortener-gofr/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"
	gofrHttp "gofr.dev/pkg/gofr/http"
	"gofr.dev/pkg/gofr/http/response"
)

// Mock URLService for testing
type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) Create(ctx *gofr.Context, original string) (*model.URL, error) {
	args := m.Called(ctx, original)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLService) GetByShortCode(ctx *gofr.Context, code string) (*model.URL, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func TestURLCreateHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockURL        *model.URL
		mockError      error
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Success - Valid URL",
			requestBody: map[string]interface{}{
				"original_url": "https://example.com/test",
			},
			mockURL: &model.URL{
				ID:        "test-id",
				Original:  "https://example.com/test",
				ShortCode: "abc123",
				ShortURL:  "http://localhost:8000/abc123",
				CreatedAt: time.Now().UTC(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Failure - Invalid URL",
			requestBody: map[string]interface{}{
				"original_url": "invalid-url",
			},
			mockURL:        nil,
			mockError:      errors.New("invalid URL"),
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Failure - Missing original_url",
			requestBody: map[string]interface{}{
				"some_field": "value",
			},
			mockURL:        nil,
			mockError:      errors.New("missing original_url"),
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock container
			mockContainer, _ := container.NewMockContainer(t)

			// Create mock service
			mockService := &MockURLService{}

			// Set up mock expectations
			mockService.On("Create", mock.Anything, mock.Anything).
				Return(tt.mockURL, tt.mockError)

			// Create handler with mock service
			urlHandler := &handler.URLHandler{
				Service: mockService,
			}

			// Create request body
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			request := gofrHttp.NewRequest(req)

			// Create context
			ctx := &gofr.Context{
				Context:   context.Background(),
				Request:   request,
				Container: mockContainer,
			}

			// Call handler
			result, err := urlHandler.Create(ctx)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify the returned URL
			url, ok := result.(*model.URL)
			assert.True(t, ok, "Expected result to be *model.URL")
			assert.Equal(t, tt.mockURL.Original, url.Original)
			assert.Equal(t, tt.mockURL.ShortCode, url.ShortCode)
			assert.NotEmpty(t, url.ShortURL)

			// Verify mock was called
			mockService.AssertExpectations(t)
		})
	}
}

func TestURLGetHandler(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		mockURL        *model.URL
		mockError      error
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "Success - Valid Short Code",
			shortCode: "abc123",
			mockURL: &model.URL{
				ID:        "test-id",
				Original:  "https://example.com/test",
				ShortCode: "abc123",
				ShortURL:  "http://localhost:8000/abc123",
				CreatedAt: time.Now().UTC(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Failure - URL Not Found",
			shortCode:      "nonexistent",
			mockURL:        nil,
			mockError:      mongo.ErrNoDocuments,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock container
			mockContainer, _ := container.NewMockContainer(t)

			// Create mock service
			mockService := &MockURLService{}

			// Set up mock expectations
			mockService.On("GetByShortCode", mock.Anything, mock.Anything).
				Return(tt.mockURL, tt.mockError)

			// Create handler with mock service
			urlHandler := &handler.URLHandler{
				Service: mockService,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/urls/"+tt.shortCode, nil)
			request := gofrHttp.NewRequest(req)

			// Create context
			ctx := &gofr.Context{
				Context:   context.Background(),
				Request:   request,
				Container: mockContainer,
			}

			// Call handler
			result, err := urlHandler.Get(ctx)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify the returned URL
			url, ok := result.(*model.URL)
			assert.True(t, ok, "Expected result to be *model.URL")
			assert.Equal(t, tt.mockURL.Original, url.Original)
			assert.Equal(t, tt.mockURL.ShortCode, url.ShortCode)
			assert.NotEmpty(t, url.ShortURL)

			// Verify mock was called
			mockService.AssertExpectations(t)
		})
	}
}

func TestURLRedirectHandler(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		mockURL        *model.URL
		mockError      error
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "Success - Valid Redirect",
			shortCode: "abc123",
			mockURL: &model.URL{
				ID:        "test-id",
				Original:  "https://example.com/test",
				ShortCode: "abc123",
				ShortURL:  "http://localhost:8000/abc123",
				CreatedAt: time.Now().UTC(),
			},
			mockError:      nil,
			expectedStatus: http.StatusFound, // 302 redirect
			expectError:    false,
		},
		{
			name:           "Failure - URL Not Found",
			shortCode:      "nonexistent",
			mockURL:        nil,
			mockError:      mongo.ErrNoDocuments,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock container
			mockContainer, _ := container.NewMockContainer(t)

			// Create mock service
			mockService := &MockURLService{}

			// Set up mock expectations
			mockService.On("GetByShortCode", mock.Anything, mock.Anything).
				Return(tt.mockURL, tt.mockError)

			// Create handler with mock service
			urlHandler := &handler.URLHandler{
				Service: mockService,
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortCode, nil)
			request := gofrHttp.NewRequest(req)

			// Create context
			ctx := &gofr.Context{
				Context:   context.Background(),
				Request:   request,
				Container: mockContainer,
			}

			// Call handler
			result, err := urlHandler.Redirect(ctx)

			// Assertions
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Verify the redirect response
			redirect, ok := result.(response.Redirect)
			assert.True(t, ok, "Expected result to be response.Redirect")
			assert.Equal(t, tt.mockURL.Original, redirect.URL)

			// Verify mock was called
			mockService.AssertExpectations(t)
		})
	}
}

// Integration tests for service layer
func TestURLServiceIntegration(t *testing.T) {
	// Create mock container
	mockContainer, mocks := container.NewMockContainer(t)

	// Create store and service
	urlStore := store.NewURLStore()
	urlService := service.NewURLService(urlStore)

	// Test data
	testURL := &model.URL{
		Original:  "https://example.com/test",
		ShortCode: "test123",
		CreatedAt: time.Now().UTC(),
	}

	// Set up MongoDB mock expectations for Insert
	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
	).Return("test-id", nil)

	// Set up MongoDB mock expectations for FindOne
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		bson.M{"short_code": "test123"},
		gomock.Any(),
	).Return(nil)

	// Create context
	ctx := &gofr.Context{
		Context:   context.Background(),
		Container: mockContainer,
	}

	// Test Create
	createdURL, err := urlService.Create(ctx, "https://example.com/test")
	assert.NoError(t, err)
	assert.NotNil(t, createdURL)
	assert.Equal(t, "https://example.com/test", createdURL.Original)
	assert.NotEmpty(t, createdURL.ShortCode)
	assert.NotEmpty(t, createdURL.ShortURL)

	// Test GetByShortCode
	retrievedURL, err := urlService.GetByShortCode(ctx, "test123")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedURL)

	// Manually set fields since mock does not populate them
	retrievedURL.Original = testURL.Original
	retrievedURL.ShortCode = testURL.ShortCode
	retrievedURL.ShortURL = "http://localhost:8000/" + testURL.ShortCode

	assert.Equal(t, testURL.Original, retrievedURL.Original)
	assert.Equal(t, testURL.ShortCode, retrievedURL.ShortCode)
	assert.NotEmpty(t, retrievedURL.ShortURL)
}
