package rest

import (
	"maps"
	"net/http"
	"time"
)

// Metrics Middleware to track http stats
func Metrics(client *Client) ClientMiddleware {
	return func(c Doer) Doer {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			start := time.Now()
			res, err := c.Do(r)

			// record metrics or errors after performing the request.
			headers := maps.Clone(r.Header)
			if err != nil {
				client.OnError(r.Method, headers, err)
			} else {
				client.OnStats(r.Host, r.Method, headers, res.StatusCode, time.Since(start))
			}

			return res, err
		})
	}
}

func generateLabelKeys(addLabelKeys ...string) []string {
	labelKeys := make([]string, len(addLabelKeys))
	for i, label := range addLabelKeys {
		labelKeys[i] = label
	}
	return labelKeys
}

func generateLabelVals(headers http.Header, addLabelVals ...string) []string {
	labelVals := make([]string, len(addLabelVals))
	for i, label := range addLabelVals {
		labelVals[i] = label
	}

	return labelVals
}

func toMilliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}
