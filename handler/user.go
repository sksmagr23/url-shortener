package handler

import (
	"github.com/sksmagr23/url-shortener-gofr/auth"
	"github.com/sksmagr23/url-shortener-gofr/service"
	"gofr.dev/pkg/gofr"
)

type UserHandler struct {
	Service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

func (h *UserHandler) Register(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}

	return h.Service.Register(ctx, service.RegisterInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
}

func (h *UserHandler) Login(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Identifier string `json:"identifier"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}

	identifier := req.Identifier
	if identifier == "" {
		identifier = req.Email
	}
	if identifier == "" {
		identifier = req.Username
	}

	return h.Service.Login(ctx, service.LoginInput{
		Identifier: identifier,
		Password:   req.Password,
	})
}

func (h *UserHandler) GetProfile(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	return h.Service.GetProfile(ctx, userID)
}

func (h *UserHandler) UpdateProfile(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}

	return h.Service.UpdateProfile(ctx, userID, service.UpdateProfileInput{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
}

func (h *UserHandler) GenerateAPIKey(ctx *gofr.Context) (interface{}, error) {
	userID, ok := auth.UserIDFromContext(ctx.Context)
	if !ok {
		return nil, service.StatusError{Code: 401, Message: "missing authenticated user"}
	}

	apiKey, err := h.Service.GenerateAPIKey(ctx, userID)
	if err != nil {
		return nil, err
	}

	return map[string]string{"api_key": apiKey}, nil
}
