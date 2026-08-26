package store_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gofr.dev/pkg/gofr"

	"github.com/sksmagr23/url-shortener-gofr/model"
	"github.com/sksmagr23/url-shortener-gofr/store"
)

func TestURLCacheGracefulNilRedis(t *testing.T) {
	cache := store.NewURLCache()
	ctx := &gofr.Context{}

	// Test GetURL when Redis is nil
	url, err := cache.GetURL(ctx, "test-code")
	assert.Error(t, err)
	assert.Nil(t, url)

	// Test SetURL when Redis is nil
	err = cache.SetURL(ctx, &model.URL{ShortCode: "test-code"}, time.Hour)
	assert.NoError(t, err)

	// Test DeleteURL when Redis is nil
	err = cache.DeleteURL(ctx, "test-code")
	assert.NoError(t, err)
}
