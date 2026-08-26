package handler

import (
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"

	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/service"
)

type URLHandler struct {
	Service service.URLService
}

func NewURLHandler(service service.URLService) *URLHandler {
	return &URLHandler{Service: service}
}

// POST /api/urls
func (h *URLHandler) Create(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	var req struct {
		OriginalURL  string     `json:"original_url"`
		CustomCode   string     `json:"custom_code"`
		Public       bool       `json:"public"`
		CustomDomain string     `json:"custom_domain"`
		ExpiresAt    *time.Time `json:"expires_at"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}
	url, err := h.Service.Create(ctx, userID, service.URLCreateInput{
		Original:     req.OriginalURL,
		CustomCode:   req.CustomCode,
		Public:       req.Public,
		CustomDomain: req.CustomDomain,
		ExpiresAt:    req.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return url, nil
}

// GET /api/urls
func (h *URLHandler) List(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	result, err := h.Service.List(ctx, userID, serviceListOptions(ctx))
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GET /api/urls/{short_code}
func (h *URLHandler) Get(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	code := ctx.PathParam("short_code")
	url, err := h.Service.GetByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	return url, nil
}

// PUT /api/urls/{short_code}
func (h *URLHandler) Update(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	var req struct {
		OriginalURL       string     `json:"original_url"`
		Public            *bool      `json:"public"`
		CustomDomain      *string    `json:"custom_domain"`
		ClearCustomDomain bool       `json:"clear_custom_domain"`
		ExpiresAt         *time.Time `json:"expires_at"`
		ClearExpiry       bool       `json:"clear_expiry"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}

	url, err := h.Service.Update(ctx, userID, ctx.PathParam("short_code"), service.URLUpdateInput{
		Original:          req.OriginalURL,
		Public:            req.Public,
		CustomDomain:      req.CustomDomain,
		ClearCustomDomain: req.ClearCustomDomain,
		ExpiresAt:         req.ExpiresAt,
		ClearExpiry:       req.ClearExpiry,
	})
	if err != nil {
		return nil, err
	}

	return url, nil
}

// DELETE /api/urls/{short_code}
func (h *URLHandler) Delete(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	if err := h.Service.Delete(ctx, userID, ctx.PathParam("short_code")); err != nil {
		return nil, err
	}

	return map[string]string{"message": "URL deleted successfully"}, nil
}

// GET /{short_code}
func (h *URLHandler) Redirect(ctx *gofr.Context) (interface{}, error) {
	code := ctx.PathParam("short_code")
	userID, _ := auth.UserIDFromContext(ctx.Context)
	url, err := h.Service.GetRedirectByShortCode(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	return response.Redirect{URL: url.Original}, nil
}

// GET /urls/{short_code}/analytics
func (h *URLHandler) GetAnalyticsSummary(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	code := ctx.PathParam("short_code")
	summary, err := h.Service.GetAnalyticsSummary(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// GET /urls/{short_code}/analytics/timeseries
func (h *URLHandler) GetAnalyticsTimeseries(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	code := ctx.PathParam("short_code")
	unit := ctx.Param("unit")
	limit, _ := strconv.Atoi(ctx.Param("limit"))

	timeseries, err := h.Service.GetAnalyticsTimeseries(ctx, userID, code, unit, limit)
	if err != nil {
		return nil, err
	}
	return timeseries, nil
}

// GET /urls/{short_code}/qr
func (h *URLHandler) GetQRCode(ctx *gofr.Context) (interface{}, error) {
	code := ctx.PathParam("short_code")
	userID, _ := auth.UserIDFromContext(ctx.Context)
	size, _ := strconv.Atoi(ctx.Param("size"))
	format := strings.ToLower(ctx.Param("format"))

	qrResp, err := h.Service.GetQRCode(ctx, userID, code, size)
	if err != nil {
		return nil, err
	}

	if format == "json" {
		return qrResp, nil
	}

	return response.File{
		Content:     qrResp.PNGBytes,
		ContentType: "image/png",
	}, nil
}

func serviceListOptions(ctx *gofr.Context) model.URLListOptions {
	page, _ := strconv.Atoi(ctx.Param("page"))
	limit, _ := strconv.Atoi(ctx.Param("limit"))

	return model.URLListOptions{
		Page:  page,
		Limit: limit,
		Sort:  ctx.Param("sort"),
		Order: ctx.Param("order"),
	}
}
