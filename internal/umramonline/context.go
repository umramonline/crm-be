package umramonline

import "context"

type contextKey struct{}

// WithToken returns a copy of ctx that carries the Umramonline Sanctum bearer token.
func WithToken(ctx context.Context, token string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKey{}, token)
}

// TokenFromContext returns the Umramonline bearer token stored in ctx, if any.
func TokenFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, _ := ctx.Value(contextKey{}).(string)

	return token
}
