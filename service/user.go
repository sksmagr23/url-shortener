package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	gofrHTTP "gofr.dev/pkg/gofr/http"
	"golang.org/x/crypto/bcrypt"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/model"
)

const tokenTTL = 24 * time.Hour

type UserRepository interface {
	Insert(ctx *gofr.Context, user *model.User) error
	FindByEmailOrUsername(ctx *gofr.Context, identifier string) (*model.User, error)
	FindByID(ctx *gofr.Context, id string) (*model.User, error)
	UpdateProfile(ctx *gofr.Context, id string, update model.UserProfileUpdate) (*model.User, error)
	AddAPIKey(ctx *gofr.Context, id, apiKey string) error
	RemoveAPIKey(ctx *gofr.Context, id, apiKey string) error
}

type UserService interface {
	Register(ctx *gofr.Context, input RegisterInput) (*model.User, error)
	Login(ctx *gofr.Context, input LoginInput) (*LoginResponse, error)
	GetProfile(ctx *gofr.Context, userID string) (*model.User, error)
	UpdateProfile(ctx *gofr.Context, userID string, input UpdateProfileInput) (*model.User, error)
	GenerateAPIKey(ctx *gofr.Context, userID string) (string, error)
	ListAPIKeys(ctx *gofr.Context, userID string) ([]string, error)
	RevokeAPIKey(ctx *gofr.Context, userID, apiKey string) error
}

type UserServiceImpl struct {
	Store     UserRepository
	JWTSecret string
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Identifier string
	Password   string
}

type UpdateProfileInput struct {
	Username string
	Email    string
	Password string
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func NewUserService(store UserRepository, jwtSecret string) UserService {
	return &UserServiceImpl{Store: store, JWTSecret: jwtSecret}
}

func (s *UserServiceImpl) Register(ctx *gofr.Context, input RegisterInput) (*model.User, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Username == "" || input.Email == "" || input.Password == "" {
		return nil, badRequest("username, email and password are required")
	}

	existing, err := s.Store.FindByEmailOrUsername(ctx, input.Email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if existing != nil {
		return nil, gofrHTTP.ErrorEntityAlreadyExist{}
	}

	existing, err = s.Store.FindByEmailOrUsername(ctx, input.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if existing != nil {
		return nil, gofrHTTP.ErrorEntityAlreadyExist{}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
		APIKeys:      []string{},
	}

	if err := s.Store.Insert(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserServiceImpl) Login(ctx *gofr.Context, input LoginInput) (*LoginResponse, error) {
	identifier := strings.TrimSpace(strings.ToLower(input.Identifier))
	if identifier == "" || input.Password == "" {
		return nil, badRequest("identifier and password are required")
	}

	user, err := s.Store.FindByEmailOrUsername(ctx, identifier)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, unauthorized("invalid credentials")
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, unauthorized("invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.Email, s.JWTSecret, tokenTTL)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token, User: user}, nil
}

func (s *UserServiceImpl) GetProfile(ctx *gofr.Context, userID string) (*model.User, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	return s.Store.FindByID(ctx, userID)
}

func (s *UserServiceImpl) UpdateProfile(ctx *gofr.Context, userID string, input UpdateProfileInput) (*model.User, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	update := model.UserProfileUpdate{
		Username: strings.TrimSpace(input.Username),
		Email:    strings.TrimSpace(strings.ToLower(input.Email)),
	}
	if input.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		update.PasswordHash = string(passwordHash)
	}

	return s.Store.UpdateProfile(ctx, userID, update)
}

func (s *UserServiceImpl) GenerateAPIKey(ctx *gofr.Context, userID string) (string, error) {
	if userID == "" {
		return "", unauthorized("missing authenticated user")
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	if err := s.Store.AddAPIKey(ctx, userID, apiKey); err != nil {
		return "", err
	}

	return apiKey, nil
}

func (s *UserServiceImpl) ListAPIKeys(ctx *gofr.Context, userID string) ([]string, error) {
	user, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user.APIKeys, nil
}

func (s *UserServiceImpl) RevokeAPIKey(ctx *gofr.Context, userID, apiKey string) error {
	if userID == "" {
		return unauthorized("missing authenticated user")
	}
	if strings.TrimSpace(apiKey) == "" {
		return badRequest("api key is required")
	}

	return s.Store.RemoveAPIKey(ctx, userID, apiKey)
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return "usk_" + hex.EncodeToString(bytes), nil
}
