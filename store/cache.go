package store

import (
	"encoding/json"
	"errors"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
)

type URLCache struct{}

func NewURLCache() *URLCache {
	return &URLCache{}
}

func (c *URLCache) GetURL(ctx *gofr.Context, code string) (*model.URL, error) {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil {
		return nil, errors.New("redis datasource is unavailable")
	}

	val, err := ctx.Redis.Get(ctx, "url:"+code).Result()
	if err != nil {
		return nil, err
	}

	var url model.URL
	if err := json.Unmarshal([]byte(val), &url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (c *URLCache) SetURL(ctx *gofr.Context, url *model.URL, ttl time.Duration) error {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil || url == nil {
		return nil
	}

	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	return ctx.Redis.Set(ctx, "url:"+url.ShortCode, string(data), ttl).Err()
}

func (c *URLCache) DeleteURL(ctx *gofr.Context, code string) error {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil {
		return nil
	}

	return ctx.Redis.Del(ctx, "url:"+code).Err()
}
