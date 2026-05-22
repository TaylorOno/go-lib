package headers

import (
	"net/http"
	"reflect"
	"testing"
)

func TestPlatformHeaders_extract(t *testing.T) {
	type options struct {
		StartsWithHeaders []string
		ExactHeaders      []string
	}

	tests := []struct {
		name  string
		given options
		when  *http.Request
		want  http.Header
	}{
		{
			name: "extract exact and starts with",
			given: options{
				StartsWithHeaders: []string{"X-Start-"},
				ExactHeaders:      []string{"X-Exact"},
			},
			when: &http.Request{
				Header: http.Header{
					"X-Start-One": []string{"val1"},
					"X-Exact":     []string{"val2"},
					"Other":       []string{"val3"},
				},
			},
			want: http.Header{
				"X-Start-One": []string{"val1"},
				"X-Exact":     []string{"val2"},
			},
		},
		{
			name: "canonicalization",
			given: options{
				StartsWithHeaders: []string{"x-start-"},
				ExactHeaders:      []string{"x-exact"},
			},
			when: &http.Request{
				Header: http.Header{
					"X-Start-One": []string{"val1"},
					"X-Exact":     []string{"val2"},
				},
			},
			want: http.Header{
				"X-Start-One": []string{"val1"},
				"X-Exact":     []string{"val2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPlatformHeaders(
				StartsWithHeaders(tt.given.StartsWithHeaders),
				ExactHeaders(tt.given.ExactHeaders),
			)
			gotCtx := h.headersToContext(tt.when)
			got := gotCtx.Value(httpHeaderKey{}).(http.Header)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extract() = %v, want %v", got, tt.want)
			}
		})
	}
}
