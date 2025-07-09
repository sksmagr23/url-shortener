package handler

import (
	"github.com/sksmagr23/url-shortener-gofr/service"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/http/response"
)

type URLHandler struct {
	Service service.URLServiceIface
}

func NewURLHandler(service service.URLServiceIface) *URLHandler {
	return &URLHandler{Service: service}
}

// POST /api/urls
func (h *URLHandler) Create(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		OriginalURL string `json:"original_url"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}
	url, err := h.Service.Create(ctx, req.OriginalURL)
	if err != nil {
		return nil, err
	}
	return url, nil
}

// GET /api/urls/{short_code}
func (h *URLHandler) Get(ctx *gofr.Context) (interface{}, error) {
	code := ctx.PathParam("short_code")
	url, err := h.Service.GetByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return url, nil
}

// GET /{short_code}
func (h *URLHandler) Redirect(ctx *gofr.Context) (interface{}, error) {
	code := ctx.PathParam("short_code")
	url, err := h.Service.GetByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return response.Redirect{URL: url.Original}, nil
}
