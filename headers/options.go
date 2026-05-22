package headers

import "net/http"

type HeaderOption func(*PlatformHeaders)

func StartsWithHeaders(headers []string) HeaderOption {
	return func(h *PlatformHeaders) {
		h.StartsWithHeaders = make([]string, len(headers))
		for i, v := range headers {
			h.StartsWithHeaders[i] = http.CanonicalHeaderKey(v)
		}
	}
}

func ExactHeaders(headers []string) HeaderOption {
	return func(h *PlatformHeaders) {
		h.ExactHeaders = make([]string, len(headers))
		for i, v := range headers {
			h.ExactHeaders[i] = http.CanonicalHeaderKey(v)
		}
	}
}
