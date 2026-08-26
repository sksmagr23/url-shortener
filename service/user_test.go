package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	"golang.org/x/crypto/bcrypt"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/service"
)

type fakeUserRepository struct {
	user        *model.User
	inserted    *model.User
	apiKey      string
	findErr     error
	insertErr   error
	updateErr   error
	apiKeyErr   error
	findByIDErr error
}

func (f *fakeUserRepository) Insert(_ *gofr.Context, user *model.User) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	user.ID = "user-1"
	f.inserted = user
	f.user = user

	return nil
}

func (f *fakeUserRepository) FindByEmailOrUsername(_ *gofr.Context, _ string) (*model.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.user == nil {
		return nil, mongo.ErrNoDocuments
	}

	return f.user, nil
}

func (f *fakeUserRepository) FindByID(_ *gofr.Context, _ string) (*model.User, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	if f.user == nil {
		return nil, mongo.ErrNoDocuments
	}

	return f.user, nil
}

func (f *fakeUserRepository) UpdateProfile(_ *gofr.Context, _ string, update model.UserProfileUpdate) (*model.User, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if update.Username != "" {
		f.user.Username = update.Username
	}
	if update.Email != "" {
		f.user.Email = update.Email
	}
	if update.PasswordHash != "" {
		f.user.PasswordHash = update.PasswordHash
	}

	return f.user, nil
}

func (f *fakeUserRepository) AddAPIKey(_ *gofr.Context, _ string, apiKey string) error {
	if f.apiKeyErr != nil {
		return f.apiKeyErr
	}
	f.apiKey = apiKey

	return nil
}

func (f *fakeUserRepository) RemoveAPIKey(_ *gofr.Context, _ string, apiKey string) error {
	if f.apiKeyErr != nil {
		return f.apiKeyErr
	}
	if f.apiKey == apiKey {
		f.apiKey = ""
	}

	return nil
}

func TestUserServiceRegister(t *testing.T) {
	repo := &fakeUserRepository{findErr: mongo.ErrNoDocuments}
	userService := service.NewUserService(repo, "test-secret")

	user, err := userService.Register(&gofr.Context{Context: context.Background()}, service.RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEqual(t, "password123", user.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")))
}

func TestUserServiceRegisterDuplicate(t *testing.T) {
	repo := &fakeUserRepository{user: &model.User{ID: "user-1"}}
	userService := service.NewUserService(repo, "test-secret")

	user, err := userService.Register(&gofr.Context{Context: context.Background()}, service.RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestUserServiceLogin(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &fakeUserRepository{user: &model.User{
		ID:           "user-1",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
	}}
	userService := service.NewUserService(repo, "test-secret")

	result, err := userService.Login(&gofr.Context{Context: context.Background()}, service.LoginInput{
		Identifier: "test@example.com",
		Password:   "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "user-1", result.User.ID)
}

func TestUserServiceLoginInvalidPassword(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &fakeUserRepository{user: &model.User{PasswordHash: string(passwordHash)}}
	userService := service.NewUserService(repo, "test-secret")

	result, err := userService.Login(&gofr.Context{Context: context.Background()}, service.LoginInput{
		Identifier: "test@example.com",
		Password:   "wrong-password",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestUserServiceUpdateProfile(t *testing.T) {
	repo := &fakeUserRepository{user: &model.User{ID: "user-1", Username: "old", Email: "old@example.com"}}
	userService := service.NewUserService(repo, "test-secret")

	user, err := userService.UpdateProfile(&gofr.Context{Context: context.Background()}, "user-1", service.UpdateProfileInput{
		Username: "new",
		Email:    "new@example.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, "new", user.Username)
	assert.Equal(t, "new@example.com", user.Email)
}

func TestUserServiceGenerateAPIKey(t *testing.T) {
	repo := &fakeUserRepository{user: &model.User{ID: "user-1"}}
	userService := service.NewUserService(repo, "test-secret")

	apiKey, err := userService.GenerateAPIKey(&gofr.Context{Context: context.Background()}, "user-1")

	assert.NoError(t, err)
	assert.NotEmpty(t, apiKey)
	assert.Equal(t, apiKey, repo.apiKey)
}

func TestUserServicePropagatesStoreError(t *testing.T) {
	repo := &fakeUserRepository{findErr: errors.New("database failed")}
	userService := service.NewUserService(repo, "test-secret")

	user, err := userService.Register(&gofr.Context{Context: context.Background()}, service.RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "database failed")
}
