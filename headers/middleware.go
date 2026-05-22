package headers

import (
	"context"
	"net/http"
	"strings"

	"github.com/taylorono/go-lib/rest"
)

// Middleware is rest.ClientMiddleware that automatically adds platform headers to the outgoing request context.
func Middleware(c rest.Doer) rest.Doer {
	return rest.ClientFunc(func(r *http.Request) (*http.Response, error) {
		ctxHeaders, ok := r.Context().Value(httpHeaderKey{}).(http.Header)
		if ok {
			for key := range ctxHeaders {
				r.Header.Set(key, ctxHeaders.Get(key))
			}
		}

		return c.Do(r)
	})
}

// PlatformHeaders is used to identify what headers are to be added to the request context
// which is used by PassToContext middleware
type PlatformHeaders struct {
	StartsWithHeaders []string
	ExactHeaders      []string
}

// NewPlatformHeaders creates a new PlatformHeaders middleware provider for a given context key.
func NewPlatformHeaders(opts ...HeaderOption) *PlatformHeaders {
	ph := &PlatformHeaders{}
	for _, opt := range opts {
		opt(ph)
	}

	return ph
}

// ContextHeaders is webserver middleware that adds header information from incoming requests to the context based on the configured settings.
func (h *PlatformHeaders) ContextHeaders(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(h.headersToContext(r)))
	}

	return fn
}

// headersToContext adds header information to the request context
func (h *PlatformHeaders) headersToContext(r *http.Request) context.Context {
	ctx := r.Context()
	httpHeaders := http.Header{}

HeaderLoop:
	for name, value := range r.Header {
		canonicalName := http.CanonicalHeaderKey(name)
		for _, exact := range h.ExactHeaders {
			exact = http.CanonicalHeaderKey(exact)
			if canonicalName == exact {
				vCopy := make([]string, len(value))
				copy(vCopy, value)
				httpHeaders[canonicalName] = vCopy
				continue HeaderLoop
			}
		}

		for _, start := range h.StartsWithHeaders {
			start = http.CanonicalHeaderKey(start)
			if strings.HasPrefix(canonicalName, start) {
				vCopy := make([]string, len(value))
				copy(vCopy, value)
				httpHeaders[canonicalName] = vCopy
				continue HeaderLoop
			}
		}
	}

	return context.WithValue(ctx, httpHeaderKey{}, httpHeaders)
}
