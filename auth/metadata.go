package auth

import "context"

type RequestMetadata struct {
	IPAddress  string
	UserAgent  string
	Referrer   string
	Browser    string
	OS         string
	DeviceType string
	Country    string
}

const metadataKey contextKey = "request_metadata"

func ContextWithMetadata(ctx context.Context, meta RequestMetadata) context.Context {
	return context.WithValue(ctx, metadataKey, meta)
}

func MetadataFromContext(ctx context.Context) (RequestMetadata, bool) {
	meta, ok := ctx.Value(metadataKey).(RequestMetadata)
	return meta, ok
}
