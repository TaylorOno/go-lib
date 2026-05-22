package headers

import (
	"context"
	"net/http"
)

type httpHeaderKey struct{}

func FromContext(ctx context.Context) http.Header {
	if h, ok := ctx.Value(httpHeaderKey{}).(http.Header); ok {
		return h
	}

	return http.Header{}
}

func WithHeader(ctx context.Context, h http.Header) context.Context {
	return context.WithValue(ctx, httpHeaderKey{}, h)
}
