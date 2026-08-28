package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
	gofrHTTP "gofr.dev/pkg/gofr/http"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

type URLServiceImpl struct {
	Store          *store.URLStore
	AnalyticsStore *store.AnalyticsStore
	Cache          *store.URLCache
	Host           string
}

func NewURLService(store *store.URLStore, analytics *store.AnalyticsStore, cache *store.URLCache, host string) URLService {
	return &URLServiceImpl{Store: store, AnalyticsStore: analytics, Cache: cache, Host: host}
}

func GenerateShortCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

type QRCodeResponse struct {
	ShortCode    string `json:"short_code"`
	ShortURL     string `json:"short_url"`
	QRCodeBase64 string `json:"qr_code_base64"`
	PNGBytes     []byte `json:"-"`
}

type URLService interface {
	Create(ctx *gofr.Context, userID string, input URLCreateInput) (*model.URL, error)
	GetByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error)
	List(ctx *gofr.Context, userID string, options model.URLListOptions) (*model.URLListResult, error)
	Update(ctx *gofr.Context, userID, code string, input URLUpdateInput) (*model.URL, error)
	Delete(ctx *gofr.Context, userID, code string) error
	GetRedirectByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error)
	GetAnalyticsSummary(ctx *gofr.Context, userID, code string) (*model.AnalyticsSummaryResponse, error)
	GetAnalyticsTimeseries(ctx *gofr.Context, userID, code, unit string, limit int) (*model.AnalyticsTimeseriesResponse, error)
	GetQRCode(ctx *gofr.Context, userID, code string, size int) (*QRCodeResponse, error)
	ListPublic(ctx *gofr.Context, options model.URLListOptions) (*model.URLListResult, error)
	ListVersions(ctx *gofr.Context, userID, code string) (*model.URLVersionHistory, error)
	RollbackVersion(ctx *gofr.Context, userID, code string, targetVersion int) (*model.URL, error)
}

type URLCreateInput struct {
	Original       string
	CustomCode     string
	Public         bool
	CustomDomain   string
	ExpiresAt      *time.Time
	IdempotencyKey string
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

	// If idempotency key is provided, check if we already processed this request
	idempotencyKey := strings.TrimSpace(input.IdempotencyKey)
	if idempotencyKey != "" {
		// 1. Fast lookup via Redis cache
		if s.Cache != nil {
			cachedURL, err := s.Cache.GetIdempotency(ctx, userID, idempotencyKey)
			if err == nil && cachedURL != nil {
				cachedURL.ShortURL = s.shortURL(cachedURL)
				return cachedURL, nil
			}
		}

		// 2. Persistent fallback via MongoDB
		if s.Store != nil {
			storedURL, err := s.Store.FindIdempotencyKey(ctx, userID, idempotencyKey)
			if err == nil && storedURL != nil {
				storedURL.ShortURL = s.shortURL(storedURL)
				if s.Cache != nil {
					_ = s.Cache.SetIdempotency(ctx, userID, idempotencyKey, storedURL, 24*time.Hour)
				}
				return storedURL, nil
			}
		}
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
		Original:       input.Original,
		ShortCode:      code,
		UserID:         userID,
		Public:         input.Public,
		CustomDomain:   domain,
		ExpiresAt:      input.ExpiresAt,
		CurrentVersion: 1,
	}
	url.ShortURL = s.shortURL(url)
	err := s.Store.Insert(ctx, url)
	if err != nil {
		return nil, err
	}

	_ = s.Store.SaveVersion(ctx, &model.URLVersion{
		URLID:        url.ID,
		ShortCode:    url.ShortCode,
		Version:      1,
		Original:     url.Original,
		CustomDomain: url.CustomDomain,
		Public:       url.Public,
		ExpiresAt:    url.ExpiresAt,
		ChangedBy:    userID,
		ChangeReason: "Initial creation",
		CreatedAt:    time.Now().UTC(),
	})

	// If idempotency key was supplied, save record for future duplicate request deduplication
	if idempotencyKey != "" {
		if s.Cache != nil {
			_ = s.Cache.SetIdempotency(ctx, userID, idempotencyKey, url, 24*time.Hour)
		}
		if s.Store != nil {
			_ = s.Store.SaveIdempotencyKey(ctx, &model.IdempotencyRecord{
				Key:       idempotencyKey,
				UserID:    userID,
				URL:       url,
				CreatedAt: time.Now().UTC(),
			})
		}
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

func (s *URLServiceImpl) ListPublic(ctx *gofr.Context, options model.URLListOptions) (*model.URLListResult, error) {
	normalizeListOptions(&options)
	urls, err := s.Store.ListPublic(ctx)
	if err != nil {
		return nil, err
	}

	sortURLs(urls, options.Sort, options.Order)

	total, err := s.Store.CountPublic(ctx)
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

	existing, err := s.Store.FindByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	newVersionNumber := existing.CurrentVersion + 1
	if newVersionNumber <= 1 {
		newVersionNumber = 2
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

	url.CurrentVersion = newVersionNumber
	_ = s.Store.UpdateOneField(ctx, userID, code, "current_version", newVersionNumber)

	_ = s.Store.SaveVersion(ctx, &model.URLVersion{
		URLID:        url.ID,
		ShortCode:    url.ShortCode,
		Version:      newVersionNumber,
		Original:     url.Original,
		CustomDomain: url.CustomDomain,
		Public:       url.Public,
		ExpiresAt:    url.ExpiresAt,
		ChangedBy:    userID,
		ChangeReason: "URL update",
		CreatedAt:    time.Now().UTC(),
	})

	if s.Cache != nil {
		_ = s.Cache.DeleteURL(ctx, code)
	}

	url.ShortURL = s.shortURL(url)

	return url, nil
}

func (s *URLServiceImpl) Delete(ctx *gofr.Context, userID, code string) error {
	if userID == "" {
		return unauthorized("missing authenticated user")
	}

	if s.Cache != nil {
		_ = s.Cache.DeleteURL(ctx, code)
	}

	return s.Store.DeleteByShortCode(ctx, userID, code)
}

func (s *URLServiceImpl) GetRedirectByShortCode(ctx *gofr.Context, userID, code string) (*model.URL, error) {
	var targetURL *model.URL

	if s.Cache != nil {
		cachedURL, err := s.Cache.GetURL(ctx, code)
		if err == nil && cachedURL != nil {
			targetURL = cachedURL
		}
	}

	if targetURL == nil {
		dbURL, err := s.Store.FindPublicByShortCode(ctx, code)
		if err != nil {
			return nil, err
		}
		targetURL = dbURL

		if s.Cache != nil {
			_ = s.Cache.SetURL(ctx, targetURL, 24*time.Hour)
		}
	}

	if targetURL.ExpiresAt != nil && time.Now().After(*targetURL.ExpiresAt) {
		return nil, gofrHTTP.ErrorEntityNotFound{Name: "short_code", Value: code}
	}
	if !targetURL.Public && targetURL.UserID != userID {
		return nil, unauthorized("private URL requires owner authentication")
	}

	targetURL.ShortURL = s.shortURL(targetURL)

	meta, ok := auth.MetadataFromContext(ctx.Context)
	if s.AnalyticsStore != nil && s.Store != nil {
		go s.recordClickAsync(ctx, targetURL.ID, code, meta, ok)
	}

	return targetURL, nil
}

func (s *URLServiceImpl) recordClickAsync(ctx *gofr.Context, urlID, code string, meta auth.RequestMetadata, metaValid bool) {
	bgCtx := &gofr.Context{
		Context:   context.Background(),
		Container: ctx.Container,
	}

	var isUnique bool
	if metaValid && meta.IPAddress != "" {
		hasClicked, _ := s.AnalyticsStore.HasIPClicked(bgCtx, code, meta.IPAddress)
		isUnique = !hasClicked

		click := &model.ClickEvent{
			URLID:      urlID,
			ShortCode:  code,
			Timestamp:  time.Now().UTC(),
			IPAddress:  meta.IPAddress,
			UserAgent:  meta.UserAgent,
			Browser:    meta.Browser,
			OS:         meta.OS,
			DeviceType: meta.DeviceType,
			Country:    meta.Country,
			Referrer:   meta.Referrer,
		}
		_ = s.AnalyticsStore.InsertClick(bgCtx, click)
	}

	_ = s.Store.IncrementClicks(bgCtx, code, isUnique)
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

func (s *URLServiceImpl) GetAnalyticsSummary(ctx *gofr.Context, userID, code string) (*model.AnalyticsSummaryResponse, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	url, err := s.Store.FindByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	browsers, _ := s.AnalyticsStore.GetFieldBreakdown(ctx, code, "browser", 10)
	osList, _ := s.AnalyticsStore.GetFieldBreakdown(ctx, code, "os", 10)
	devices, _ := s.AnalyticsStore.GetFieldBreakdown(ctx, code, "device_type", 10)
	countries, _ := s.AnalyticsStore.GetFieldBreakdown(ctx, code, "country", 10)
	referrers, _ := s.AnalyticsStore.GetFieldBreakdown(ctx, code, "referrer", 10)

	return &model.AnalyticsSummaryResponse{
		ShortCode:    code,
		TotalClicks:  url.TotalClicks,
		UniqueClicks: url.UniqueClicks,
		Browsers:     browsers,
		OS:           osList,
		Devices:      devices,
		Countries:    countries,
		Referrers:    referrers,
	}, nil
}

func (s *URLServiceImpl) GetAnalyticsTimeseries(ctx *gofr.Context, userID, code, unit string, limit int) (*model.AnalyticsTimeseriesResponse, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	_, err := s.Store.FindByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	if unit != "hour" && unit != "day" {
		unit = "day"
	}

	ts, err := s.AnalyticsStore.GetTimeseries(ctx, code, unit, limit)
	if err != nil {
		return nil, err
	}

	return &model.AnalyticsTimeseriesResponse{
		ShortCode:  code,
		Unit:       unit,
		Timeseries: ts,
	}, nil
}

func (s *URLServiceImpl) GetQRCode(ctx *gofr.Context, userID, code string, size int) (*QRCodeResponse, error) {
	if size <= 0 {
		size = 256
	}
	if size > 1024 {
		size = 1024
	}

	url, err := s.Store.FindPublicByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if !url.Public && url.UserID != userID {
		return nil, unauthorized("private URL requires owner authentication")
	}

	shortURL := s.shortURL(url)
	pngBytes, err := qrcode.Encode(shortURL, qrcode.Medium, size)
	if err != nil {
		return nil, badRequest("failed to generate QR code")
	}

	base64Str := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)

	return &QRCodeResponse{
		ShortCode:    code,
		ShortURL:     shortURL,
		QRCodeBase64: base64Str,
		PNGBytes:     pngBytes,
	}, nil
}

func (s *URLServiceImpl) ListVersions(ctx *gofr.Context, userID, code string) (*model.URLVersionHistory, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	url, err := s.Store.FindByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	versions, err := s.Store.GetVersionsByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	for _, v := range versions {
		if v.Version == url.CurrentVersion {
			v.IsCurrent = true
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	return &model.URLVersionHistory{
		ShortCode:      code,
		CurrentVersion: url.CurrentVersion,
		Versions:       versions,
	}, nil
}

func (s *URLServiceImpl) RollbackVersion(ctx *gofr.Context, userID, code string, targetVersion int) (*model.URL, error) {
	if userID == "" {
		return nil, unauthorized("missing authenticated user")
	}

	url, err := s.Store.FindByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}

	targetVer, err := s.Store.GetVersionByNumber(ctx, code, targetVersion)
	if err != nil {
		return nil, badRequest(fmt.Sprintf("version %d not found for code '%s'", targetVersion, code))
	}

	newVersionNumber := url.CurrentVersion + 1
	if newVersionNumber <= targetVersion {
		newVersionNumber = targetVersion + 1
	}

	updatedURL, err := s.Store.UpdateByShortCode(ctx, userID, code, model.URLUpdate{
		Original:     targetVer.Original,
		Public:       &targetVer.Public,
		CustomDomain: &targetVer.CustomDomain,
		ExpiresAt:    targetVer.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	updatedURL.CurrentVersion = newVersionNumber
	_ = s.Store.UpdateOneField(ctx, userID, code, "current_version", newVersionNumber)

	_ = s.Store.SaveVersion(ctx, &model.URLVersion{
		URLID:        updatedURL.ID,
		ShortCode:    code,
		Version:      newVersionNumber,
		Original:     targetVer.Original,
		CustomDomain: targetVer.CustomDomain,
		Public:       targetVer.Public,
		ExpiresAt:    targetVer.ExpiresAt,
		ChangedBy:    userID,
		ChangeReason: fmt.Sprintf("Rolled back to version %d", targetVersion),
		CreatedAt:    time.Now().UTC(),
	})

	if s.Cache != nil {
		_ = s.Cache.DeleteURL(ctx, code)
	}

	updatedURL.ShortURL = s.shortURL(updatedURL)

	return updatedURL, nil
}
