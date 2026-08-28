package auth

import "context"

type RequestMetadata struct {
	IPAddress      string
	UserAgent      string
	Referrer       string
	Browser        string
	OS             string
	DeviceType     string
	Country        string
	IdempotencyKey string
}

const (
	metadataKey       contextKey = "request_metadata"
	idempotencyKeyCtx contextKey = "idempotency_key"
)

func ContextWithMetadata(ctx context.Context, meta RequestMetadata) context.Context {
	return context.WithValue(ctx, metadataKey, meta)
}

func MetadataFromContext(ctx context.Context) (RequestMetadata, bool) {
	meta, ok := ctx.Value(metadataKey).(RequestMetadata)
	return meta, ok
}

func ContextWithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtx, key)
}

func IdempotencyKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if key, ok := ctx.Value(idempotencyKeyCtx).(string); ok && key != "" {
		return key
	}
	if meta, ok := MetadataFromContext(ctx); ok && meta.IdempotencyKey != "" {
		return meta.IdempotencyKey
	}
	return ""
}
