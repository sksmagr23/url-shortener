package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"gofr.dev/pkg/gofr"
	"golang.org/x/crypto/bcrypt"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

type UserService interface {
	Register(ctx *gofr.Context, username, email, password string) (*model.User, error)
	Login(ctx *gofr.Context, email, password string) (string, *model.User, error)
	GetProfile(ctx *gofr.Context, userID string) (*model.User, error)
	UpdateProfile(ctx *gofr.Context, userID, username, email string) error
	GenerateAPIKey(ctx *gofr.Context, userID string) (string, error)
}

type UserServiceImpl struct {
	Store     *store.UserStore
	JWTSecret string
}

func NewUserService(store *store.UserStore, jwtSecret string) UserService {
	return &UserServiceImpl{Store: store, JWTSecret: jwtSecret}
}

func (s *UserServiceImpl) Register(
	ctx *gofr.Context,
	username, email, password string,
) (*model.User, error) {
	if user, _ := s.Store.FindByEmail(ctx, email); user != nil {
		return nil, errors.New("email already registered")
	}
	if user, _ := s.Store.FindByUsername(ctx, username); user != nil {
		return nil, errors.New("username already taken")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
	}
	err = s.Store.Insert(ctx, user)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *UserServiceImpl) Login(
	ctx *gofr.Context,
	email, password string,
) (string, *model.User, error) {
	user, err := s.Store.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}
	if !CheckPassword(user.PasswordHash, password) {
		return "", nil, errors.New("invalid email or password")
	}
	token, err := GenerateJWT(user, s.JWTSecret)
	if err != nil {
		return "", nil, err
	}
	user.PasswordHash = ""
	return token, user, nil
}

func (s *UserServiceImpl) GetProfile(ctx *gofr.Context, userID string) (*model.User, error) {
	user, err := s.Store.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *UserServiceImpl) UpdateProfile(ctx *gofr.Context, userID, username, email string) error {
	update := bson.M{}
	if username != "" {
		update["username"] = username
	}
	if email != "" {
		update["email"] = email
	}
	if len(update) == 0 {
		return errors.New("no fields to update")
	}
	return s.Store.Update(ctx, userID, update)
}

func (s *UserServiceImpl) GenerateAPIKey(ctx *gofr.Context, userID string) (string, error) {
	apiKey := GenerateRandomAPIKey(32)
	err := s.Store.Update(ctx, userID, bson.M{"api_key": apiKey})
	if err != nil {
		return "", err
	}
	return apiKey, nil
}

// --- Helpers ---

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateJWT(user *model.User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func ValidateJWT(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("user_id missing in token")
	}
	return userID, nil
}

func GenerateRandomAPIKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
