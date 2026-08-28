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

func (c *URLCache) GetURL(ctx *gofr.Context, code string) (result *model.URL, err error) {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil {
		return nil, errors.New("redis datasource is unavailable")
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("redis datasource is unavailable")
			result = nil
		}
	}()

	val, getErr := ctx.Redis.Get(ctx, "url:"+code).Result()
	if getErr != nil {
		return nil, getErr
	}

	var url model.URL
	if unmarshalErr := json.Unmarshal([]byte(val), &url); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &url, nil
}

func (c *URLCache) SetURL(ctx *gofr.Context, url *model.URL, ttl time.Duration) (err error) {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil || url == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = nil
		}
	}()

	data, marshalErr := json.Marshal(url)
	if marshalErr != nil {
		return marshalErr
	}

	return ctx.Redis.Set(ctx, "url:"+url.ShortCode, string(data), ttl).Err()
}

func (c *URLCache) DeleteURL(ctx *gofr.Context, code string) (err error) {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = nil
		}
	}()

	return ctx.Redis.Del(ctx, "url:"+code).Err()
}

func (c *URLCache) GetIdempotency(ctx *gofr.Context, userID, key string) (result *model.URL, err error) {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil || userID == "" || key == "" {
		return nil, errors.New("redis datasource is unavailable")
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("redis datasource is unavailable")
			result = nil
		}
	}()

	val, getErr := ctx.Redis.Get(ctx, "idempotency:"+userID+":"+key).Result()
	if getErr != nil {
		return nil, getErr
	}

	var url model.URL
	if unmarshalErr := json.Unmarshal([]byte(val), &url); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &url, nil
}

func (c *URLCache) SetIdempotency(ctx *gofr.Context, userID, key string, url *model.URL, ttl time.Duration) (err error) {
	if ctx == nil || ctx.Container == nil || ctx.Redis == nil || url == nil || userID == "" || key == "" {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = nil
		}
	}()

	data, marshalErr := json.Marshal(url)
	if marshalErr != nil {
		return marshalErr
	}

	return ctx.Redis.Set(ctx, "idempotency:"+userID+":"+key, string(data), ttl).Err()
}
