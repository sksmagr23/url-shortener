package handler

import (
	"errors"

	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/service"
)

type UserHandler struct {
	Service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

// POST /users/register
func (h *UserHandler) Register(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("username, email, and password are required")
	}
	user, err := h.Service.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"message": "User registered successfully",
			"user":    user,
		},
	}, nil
}

// POST /users/login
func (h *UserHandler) Login(ctx *gofr.Context) (interface{}, error) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}
	token, user, err := h.Service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": map[string]interface{}{
			"token": token,
			"user":  user,
		},
	}, nil
}

// GET /users/profile (JWT required)
func (h *UserHandler) Profile(ctx *gofr.Context) (interface{}, error) {
	claims, ok := ctx.Request.Context().Value("claims").(map[string]interface{})
if !ok {
    return nil, errors.New("invalid or missing JWT claims")
}
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("user_id not found in token")
	}
	user, err := h.Service.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": user}, nil
}

// PUT /users/profile (JWT required)
func (h *UserHandler) UpdateProfile(ctx *gofr.Context) (interface{}, error) {
	claims, ok := ctx.Request.Context().Value("claims").(map[string]interface{})
if !ok {
    return nil, errors.New("invalid or missing JWT claims")
}
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("user_id not found in token")
	}
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := ctx.Bind(&req); err != nil {
		return nil, err
	}
	if req.Username == "" && req.Email == "" {
		return nil, errors.New("at least one field (username or email) is required")
	}
	err := h.Service.UpdateProfile(ctx, userID, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"data": map[string]string{"message": "Profile updated successfully"},
	}, nil
}

// POST /users/api-key (JWT required)
func (h *UserHandler) GenerateAPIKey(ctx *gofr.Context) (interface{}, error) {
	claims, ok := ctx.Request.Context().Value("claims").(map[string]interface{})
if !ok {
    return nil, errors.New("invalid or missing JWT claims")
}
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("user_id not found in token")
	}
	apiKey, err := h.Service.GenerateAPIKey(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": map[string]string{"api_key": apiKey}}, nil
}
