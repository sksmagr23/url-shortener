package service

import (
	"errors"
	"math/rand"
	"os"
	"strings"

	"github.com/sksmagr23/url-shortener-gofr/internal/model"
	"github.com/sksmagr23/url-shortener-gofr/internal/store"
	"gofr.dev/pkg/gofr"
)

type URLService struct {
	Store *store.URLStore
}

func NewURLService(store *store.URLStore) *URLService {
	return &URLService{Store: store}
}

func GenerateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *URLService) Create(ctx *gofr.Context, original string) (*model.URL, error) {
	if !strings.HasPrefix(original, "http://") && !strings.HasPrefix(original, "https://") {
		return nil, errors.New("invalid URL")
	}
	code := GenerateShortCode(6)
	host := os.Getenv("SHORT_URL_HOST")
	url := &model.URL{
		Original:  original,
		ShortCode: code,
	}
	url.ShortURL = host + code
	err := s.Store.Insert(ctx, url)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (s *URLService) GetByShortCode(ctx *gofr.Context, code string) (*model.URL, error) {
	url, err := s.Store.FindByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}
	host := os.Getenv("SHORT_URL_HOST")
	if host == "" {
		host = "https://sksmagr23/"
	}
	url.ShortURL = host + url.ShortCode
	return url, nil
}
