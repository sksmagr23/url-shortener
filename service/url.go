package service

import (
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	gofrHTTP "gofr.dev/pkg/gofr/http"

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
	Create(ctx *gofr.Context, userID string, input URLCreateInput) (*model.URL, error)
	GetByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error)
	List(ctx *gofr.Context, userID string, options model.URLListOptions) (*model.URLListResult, error)
	Update(ctx *gofr.Context, userID, code string, input URLUpdateInput) (*model.URL, error)
	Delete(ctx *gofr.Context, userID, code string) error
	GetRedirectByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error)
}

type URLCreateInput struct {
	Original     string
	CustomCode   string
	Public       bool
	CustomDomain string
	ExpiresAt    *time.Time
}

type URLUpdateInput struct {
	Original          string
	Public            *bool
	CustomDomain      *string
	ClearCustomDomain bool
	ExpiresAt         *time.Time
	ClearExpiry       bool
}

func (s *URLServiceImpl) Create(ctx *gofr.Context, userID string, input URLCreateInput) (*model.URL, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}
	if !isHTTPURL(input.Original) {
		return nil, badRequest("invalid URL")
	}

	domain := normalizeDomain(input.CustomDomain)
	if domain != "" && !isValidDomain(domain) {
		return nil, badRequest("invalid custom domain")
	}

	code := strings.TrimSpace(input.CustomCode)
	if code == "" {
		code = GenerateShortCode(6)
	}
	if !isValidShortCode(code) {
		return nil, badRequest("invalid short code")
	}
	if err := s.ensureShortCodeAvailable(ctx, code); err != nil {
		return nil, err
	}

	url := &model.URL{
		Original:     input.Original,
		ShortCode:    code,
		UserID:       userID,
		Public:       input.Public,
		CustomDomain: domain,
		ExpiresAt:    input.ExpiresAt,
	}
	url.ShortURL = s.shortURL(url)
	err := s.Store.Insert(ctx, url)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func (s *URLServiceImpl) GetByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	url, err := s.Store.FindByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	url.ShortURL = s.shortURL(url)
	return url, nil
}

func (s *URLServiceImpl) List(ctx *gofr.Context, userID string, options model.URLListOptions) (*model.URLListResult, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	normalizeListOptions(&options)
	urls, err := s.Store.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortURLs(urls, options.Sort, options.Order)

	total, err := s.Store.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	start := (options.Page - 1) * options.Limit
	end := start + options.Limit
	if start > len(urls) {
		start = len(urls)
	}
	if end > len(urls) {
		end = len(urls)
	}

	for _, url := range urls {
		url.ShortURL = s.shortURL(url)
	}

	return &model.URLListResult{
		URLs: urls[start:end],
		Pagination: model.Pagination{
			Page:       options.Page,
			Limit:      options.Limit,
			Total:      total,
			TotalPages: totalPages(total, options.Limit),
		},
	}, nil
}

func (s *URLServiceImpl) Update(ctx *gofr.Context, userID, code string, input URLUpdateInput) (*model.URL, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}
	if input.Original != "" && !isHTTPURL(input.Original) {
		return nil, badRequest("invalid URL")
	}

	var customDomainPtr *string
	if input.CustomDomain != nil {
		norm := normalizeDomain(*input.CustomDomain)
		if norm != "" && !isValidDomain(norm) {
			return nil, badRequest("invalid custom domain")
		}
		customDomainPtr = &norm
	}

	url, err := s.Store.UpdateByShortCode(ctx, userID, code, model.URLUpdate{
		Original:          input.Original,
		Public:            input.Public,
		CustomDomain:      customDomainPtr,
		ClearCustomDomain: input.ClearCustomDomain,
		ExpiresAt:         input.ExpiresAt,
		ClearExpiry:       input.ClearExpiry,
	})
	if err != nil {
		return nil, err
	}

	url.ShortURL = s.shortURL(url)

	return url, nil
}

func (s *URLServiceImpl) Delete(ctx *gofr.Context, userID, code string) error {
	if userID == "" {
		return unauthorized("missing authenticated user")
	}

	return s.Store.DeleteByShortCode(ctx, userID, code)
}

func (s *URLServiceImpl) GetRedirectByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	url, err := s.Store.FindPublicByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, gofrHTTP.ErrorEntityNotFound{Name: "short_code", Value: code}
	}
	if !url.Public && url.UserID != userID {
		return nil, unauthorized("private URL requires owner authentication")
	}

	url.ShortURL = s.shortURL(url)
	_ = s.Store.IncrementClicks(ctx, code)

	return url, nil
}

func (s *URLServiceImpl) ensureShortCodeAvailable(ctx *gofr.Context, code string) error {
	_, err := s.Store.FindPublicByShortCode(ctx, code)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if err != nil {
		return err
	}

	return gofrHTTP.ErrorEntityAlreadyExist{}
}

func (s *URLServiceImpl) shortURL(url *model.URL) string {
	if url.CustomDomain != "" {
		domain := strings.TrimRight(url.CustomDomain, "/")
		if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
			domain = "https://" + domain
		}
		return domain + "/" + url.ShortCode
	}

	return strings.TrimRight(s.Host, "/") + "/" + url.ShortCode
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return ""
	}
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimRight(domain, "/")
	return domain
}

func isValidDomain(domain string) bool {
	if domain == "" {
		return true
	}
	if strings.Contains(domain, " ") || strings.Contains(domain, "/") || strings.Contains(domain, "?") || strings.Contains(domain, "#") {
		return false
	}
	return true
}

func isHTTPURL(original string) bool {
	return strings.HasPrefix(original, "http://") || strings.HasPrefix(original, "https://")
}

func isValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 64 {
		return false
	}

	for _, char := range code {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}

		return false
	}

	return true
}

func normalizeListOptions(options *model.URLListOptions) {
	if options.Page < 1 {
		options.Page = 1
	}
	if options.Limit < 1 || options.Limit > 100 {
		options.Limit = 10
	}
	if options.Sort == "" {
		options.Sort = "created_at"
	}
	if options.Order == "" {
		options.Order = "desc"
	}
}

func sortURLs(urls []*model.URL, field, order string) {
	desc := strings.EqualFold(order, "desc")

	sort.SliceStable(urls, func(i, j int) bool {
		var less bool
		switch field {
		case "short_code":
			less = urls[i].ShortCode < urls[j].ShortCode
		case "total_clicks":
			less = urls[i].TotalClicks < urls[j].TotalClicks
		default:
			less = urls[i].CreatedAt.Before(urls[j].CreatedAt)
		}
		if desc {
			return !less
		}

		return less
	})
}

func totalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}

	return int((total + int64(limit) - 1) / int64(limit))
}
