package contextutil

import "context"

// Menggunakan unexported type untuk keamanan context key
type contextKey string

const requestIDKey contextKey = "request_id"

// GetRequestID mengambil Request ID dari context
func GetRequestID(ctx context.Context) string {
	if rid, ok := ctx.Value(requestIDKey).(string); ok {
		return rid
	}
	// Jika tidak ada di context, coba cek apakah ctx itu gin.Context
	// Namun idealnya propagasi sudah dilakukan di middleware.
	return ""
}

// WithRequestID digunakan untuk menginjeksi ID ke context (berguna untuk Unit Test)
func WithRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIDKey, rid)
}

// GetKey didefinisikan jika middleware perlu tahu key mentahnya
func GetKey() string {
	return string(requestIDKey)
}
