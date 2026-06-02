package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/container"
	gofrHttp "gofr.dev/pkg/gofr/http"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/handler"
	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/service"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx *gofr.Context, input service.RegisterInput) (*model.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) Login(ctx *gofr.Context, input service.LoginInput) (*service.LoginResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetProfile(ctx *gofr.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx *gofr.Context, userID string, input service.UpdateProfileInput) (*model.User, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GenerateAPIKey(ctx *gofr.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)

	return args.String(0), args.Error(1)
}

func TestUserRegisterHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockUserService{}
	expectedUser := &model.User{
		ID:        "user-1",
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now().UTC(),
	}
	mockService.On("Register", mock.Anything, mock.MatchedBy(func(input service.RegisterInput) bool {
		return input.Username == "testuser" && input.Email == "test@example.com" && input.Password == "password123"
	})).Return(expectedUser, nil)

	userHandler := handler.NewUserHandler(mockService)
	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	result, err := userHandler.Register(&gofr.Context{
		Context:   context.Background(),
		Request:   gofrHttp.NewRequest(req),
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockService.AssertExpectations(t)
}

func TestUserLoginHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockUserService{}
	expected := &service.LoginResponse{Token: "jwt-token", User: &model.User{ID: "user-1"}}
	mockService.On("Login", mock.Anything, service.LoginInput{
		Identifier: "test@example.com",
		Password:   "password123",
	}).Return(expected, nil)

	userHandler := handler.NewUserHandler(mockService)
	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	result, err := userHandler.Login(&gofr.Context{
		Context:   context.Background(),
		Request:   gofrHttp.NewRequest(req),
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockService.AssertExpectations(t)
}

func TestUserGetProfileHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockUserService{}
	expectedUser := &model.User{ID: "user-1", Username: "testuser"}
	mockService.On("GetProfile", mock.Anything, "user-1").Return(expectedUser, nil)

	userHandler := handler.NewUserHandler(mockService)
	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)

	result, err := userHandler.GetProfile(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Request:   gofrHttp.NewRequest(req),
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockService.AssertExpectations(t)
}

func TestUserGenerateAPIKeyHandler(t *testing.T) {
	mockContainer, _ := container.NewMockContainer(t)
	mockService := &MockUserService{}
	mockService.On("GenerateAPIKey", mock.Anything, "user-1").Return("usk_test", nil)

	userHandler := handler.NewUserHandler(mockService)
	req := httptest.NewRequest(http.MethodPost, "/users/api-key", nil)

	result, err := userHandler.GenerateAPIKey(&gofr.Context{
		Context:   auth.ContextWithUserID(context.Background(), "user-1"),
		Request:   gofrHttp.NewRequest(req),
		Container: mockContainer,
	})

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"api_key": "usk_test"}, result)
	mockService.AssertExpectations(t)
}
