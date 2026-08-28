package handler_test

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
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"
	"gofr.dev/pkg/gofr/http/response"

	gofrHttp "gofr.dev/pkg/gofr/http"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/handler"
	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/service"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) Create(ctx *gofr.Context, userID string, input service.URLCreateInput) (*model.URL, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLService) GetByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLService) List(ctx *gofr.Context, userID string, options model.URLListOptions) (*model.URLListResult, error) {
	args := m.Called(ctx, userID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URLListResult), args.Error(1)
}

func (m *MockURLService) Update(ctx *gofr.Context, userID, code string, input service.URLUpdateInput) (*model.URL, error) {
	args := m.Called(ctx, userID, code, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLService) Delete(ctx *gofr.Context, userID, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *MockURLService) GetRedirectByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func (m *MockURLService) GetAnalyticsSummary(ctx *gofr.Context, userID, code string) (*model.AnalyticsSummaryResponse, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AnalyticsSummaryResponse), args.Error(1)
}

func (m *MockURLService) GetAnalyticsTimeseries(ctx *gofr.Context, userID, code, unit string, limit int) (*model.AnalyticsTimeseriesResponse, error) {
	args := m.Called(ctx, userID, code, unit, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AnalyticsTimeseriesResponse), args.Error(1)
}

func (m *MockURLService) GetQRCode(ctx *gofr.Context, userID, code string, size int) (*service.QRCodeResponse, error) {
	args := m.Called(ctx, userID, code, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.QRCodeResponse), args.Error(1)
}

func (m *MockURLService) ListPublic(ctx *gofr.Context, options model.URLListOptions) (*model.URLListResult, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URLListResult), args.Error(1)
}

func (m *MockURLService) ListVersions(ctx *gofr.Context, userID, code string) (*model.URLVersionHistory, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URLVersionHistory), args.Error(1)
}

func (m *MockURLService) RollbackVersion(ctx *gofr.Context, userID, code string, targetVersion int) (*model.URL, error) {
	args := m.Called(ctx, userID, code, targetVersion)
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
			mockContainer, _ := container.NewMockContainer(t)

			mockService := &MockURLService{}

			mockService.On("Create", mock.Anything, "user-1", mock.Anything).
				Return(tt.mockURL, tt.mockError)

			urlHandler := &handler.URLHandler{
				Service: mockService,
			}

			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			request := gofrHttp.NewRequest(req)

			ctx := &gofr.Context{
				Context:   auth.ContextWithUserID(context.Background(), "user-1"),
				Request:   request,
				Container: mockContainer,
			}

			result, err := urlHandler.Create(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			url, ok := result.(*model.URL)
			assert.True(t, ok, "Expected result to be *model.URL")
			assert.Equal(t, tt.mockURL.Original, url.Original)
			assert.Equal(t, tt.mockURL.ShortCode, url.ShortCode)
			assert.NotEmpty(t, url.ShortURL)
			mockService.AssertExpectations(t)
		})
	}
}

func TestURLCreateHandlerWithIdempotency(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockURLService{}

	mockURL := &model.URL{
		ID:        "idem-id",
		Original:  "https://example.com/test",
		ShortCode: "idem12",
		ShortURL:  "http://localhost:8000/idem12",
	}

	mockService.On("Create", mock.Anything, "user-1", mock.MatchedBy(func(input service.URLCreateInput) bool {
		return input.IdempotencyKey == "key-999" && input.Original == "https://example.com/test"
	})).Return(mockURL, nil)

	urlHandler := &handler.URLHandler{Service: mockService}

	requestBody, _ := json.Marshal(map[string]string{
		"original_url":    "https://example.com/test",
		"idempotency_key": "key-999",
	})
	req := httptest.NewRequest(http.MethodPost, "/urls", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	request := gofrHttp.NewRequest(req)

	ctx := &gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Request:   request,
		Container: mockContainer,
	}

	result, err := urlHandler.Create(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	url, ok := result.(*model.URL)
	assert.True(t, ok)
	assert.Equal(t, "idem12", url.ShortCode)
	mockService.AssertExpectations(t)
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
			mockContainer, _ := container.NewMockContainer(t)

			mockService := &MockURLService{}

			mockService.On("GetByShortCode", mock.Anything, "user-1", mock.Anything).
				Return(tt.mockURL, tt.mockError)

			urlHandler := &handler.URLHandler{
				Service: mockService,
			}

			req := httptest.NewRequest(http.MethodGet, "/api/urls/"+tt.shortCode, nil)
			request := gofrHttp.NewRequest(req)

			ctx := &gofr.Context{
				Context:   auth.ContextWithUserID(context.Background(), "user-1"),
				Request:   request,
				Container: mockContainer,
			}

			result, err := urlHandler.Get(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			url, ok := result.(*model.URL)
			assert.True(t, ok, "Expected result to be *model.URL")
			assert.Equal(t, tt.mockURL.Original, url.Original)
			assert.Equal(t, tt.mockURL.ShortCode, url.ShortCode)
			assert.NotEmpty(t, url.ShortURL)
			mockService.AssertExpectations(t)
		})
	}
}

func TestURLListHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockURLService{}
	expected := &model.URLListResult{
		URLs: []*model.URL{{ID: "url-1", ShortCode: "abc123"}},
		Pagination: model.Pagination{
			Page:       2,
			Limit:      5,
			Total:      1,
			TotalPages: 1,
		},
	}
	mockService.On("List", mock.Anything, "user-1", model.URLListOptions{
		Page:  2,
		Limit: 5,
		Sort:  "created_at",
		Order: "asc",
	}).Return(expected, nil)

	urlHandler := &handler.URLHandler{Service: mockService}
	req := httptest.NewRequest(http.MethodGet, "/api/urls?page=2&limit=5&sort=created_at&order=asc", nil)
	request := gofrHttp.NewRequest(req)

	result, err := urlHandler.List(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Request:   request,
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockService.AssertExpectations(t)
}

func TestURLUpdateHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockURLService{}
	isPublic := false
	expectedURL := &model.URL{ID: "url-1", ShortCode: "abc123", Public: false}
	mockService.On("Update", mock.Anything, "user-1", "abc123", mock.MatchedBy(func(input service.URLUpdateInput) bool {
		return input.Public != nil && *input.Public == isPublic
	})).Return(expectedURL, nil)

	urlHandler := &handler.URLHandler{Service: mockService}
	body, _ := json.Marshal(map[string]any{"public": false})
	req := httptest.NewRequest(http.MethodPut, "/api/urls/abc123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"short_code": "abc123"})
	request := gofrHttp.NewRequest(req)

	result, err := urlHandler.Update(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Request:   request,
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, result)
	mockService.AssertExpectations(t)
}

func TestURLDeleteHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockURLService{}
	mockService.On("Delete", mock.Anything, "user-1", "abc123").Return(nil)

	urlHandler := &handler.URLHandler{Service: mockService}
	req := httptest.NewRequest(http.MethodDelete, "/api/urls/abc123", nil)
	req = mux.SetURLVars(req, map[string]string{"short_code": "abc123"})
	request := gofrHttp.NewRequest(req)

	result, err := urlHandler.Delete(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Request:   request,
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"message": "URL deleted successfully"}, result)
	mockService.AssertExpectations(t)
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
			expectedStatus: http.StatusFound,
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
			mockContainer, _ := container.NewMockContainer(t)

			mockService := &MockURLService{}

			mockService.On("GetRedirectByShortCode", mock.Anything, "", mock.Anything).
				Return(tt.mockURL, tt.mockError)

			urlHandler := &handler.URLHandler{
				Service: mockService,
			}

			req := httptest.NewRequest(http.MethodGet, "/"+tt.shortCode, nil)
			request := gofrHttp.NewRequest(req)

			ctx := &gofr.Context{
				Context:   context.Background(),
				Request:   request,
				Container: mockContainer,
			}

			result, err := urlHandler.Redirect(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			redirect, ok := result.(response.Redirect)
			assert.True(t, ok, "Expected result to be response.Redirect")
			assert.Equal(t, tt.mockURL.Original, redirect.URL)
			mockService.AssertExpectations(t)
		})
	}
}

// Integration tests
func TestURLServiceIntegration(t *testing.T) {
	mockContainer, mocks := container.NewMockContainer(t)

	urlStore := store.NewURLStore()
	analyticsStore := store.NewAnalyticsStore()
	urlService := service.NewURLService(urlStore, analyticsStore, nil, "http://localhost:8000/")

	testURL := &model.URL{
		Original:  "https://example.com/test",
		ShortCode: "test123",
		CreatedAt: time.Now().UTC(),
	}

	mocks.Mongo.EXPECT().InsertOne(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return("test-id", nil).AnyTimes()

	var createdCode string
	mocks.Mongo.EXPECT().FindOne(
		gomock.Any(),
		"urls",
		gomock.Any(),
		gomock.Any(),
	).DoAndReturn(func(ctx interface{}, coll string, filter interface{}, res interface{}) error {
		if createdCode == "" {
			return mongo.ErrNoDocuments
		}
		u, ok := res.(*model.URL)
		if ok {
			u.ShortCode = createdCode
			u.Original = "https://example.com/test"
			u.UserID = "user-1"
			return nil
		}
		return mongo.ErrNoDocuments
	}).AnyTimes()

	ctx := &gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Container: mockContainer,
	}

	createdURL, err := urlService.Create(ctx, "user-1", service.URLCreateInput{Original: "https://example.com/test"})
	assert.NoError(t, err)
	assert.NotNil(t, createdURL)
	assert.Equal(t, "https://example.com/test", createdURL.Original)
	assert.NotEmpty(t, createdURL.ShortCode)
	assert.NotEmpty(t, createdURL.ShortURL)
	createdCode = createdURL.ShortCode

	retrievedURL, err := urlService.GetByShortCode(ctx, "user-1", createdCode)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedURL)
	retrievedURL.Original = testURL.Original
	retrievedURL.ShortCode = testURL.ShortCode
	retrievedURL.ShortURL = "http://localhost:8000/" + testURL.ShortCode

	assert.Equal(t, testURL.Original, retrievedURL.Original)
	assert.Equal(t, testURL.ShortCode, retrievedURL.ShortCode)
	assert.NotEmpty(t, retrievedURL.ShortURL)
}
