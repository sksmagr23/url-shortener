package service

import (
	"errors"
	"math/rand"
	"strings"

	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

type URLServiceImpl struct {
	Store *store.URLStore
	Host  string
}

func NewURLService(store *store.URLStore, host string) URLService {
	return &URLServiceImpl{Store: store, Host: host}
}

func GenerateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

type URLService interface {
	Create(ctx *gofr.Context, original string) (*model.URL, error)
	GetByShortCode(ctx *gofr.Context, code string) (*model.URL, error)
}

func (s *URLServiceImpl) Create(ctx *gofr.Context, original string) (*model.URL, error) {
	if !strings.HasPrefix(original, "http://") && !strings.HasPrefix(original, "https://") {
		return nil, errors.New("invalid URL")
	}
	code := GenerateShortCode(6)
	url := &model.URL{
		Original:  original,
		ShortCode: code,
	}
	url.ShortURL = s.Host + code
	err := s.Store.Insert(ctx, url)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (s *URLServiceImpl) GetByShortCode(ctx *gofr.Context, code string) (*model.URL, error) {
	url, err := s.Store.FindByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}
	url.ShortURL = s.Host + url.ShortCode
	return url, nil
}
