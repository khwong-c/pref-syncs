package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
)

// LocalTransport is intent to route local request through direct handler calls, instead of via network layers.
type LocalTransport struct {
	DefaultTransport http.RoundTripper
	LocalRoutes      []string
	Handler          http.Handler
}

func (r *LocalTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, localRoute := range r.LocalRoutes {
		if req.URL.Host == localRoute || strings.HasPrefix(req.URL.String(), localRoute) {
			rec := httptest.NewRecorder()

			ctx := context.Background() // Acting like a new request
			if deadline, hasDeadline := req.Context().Deadline(); hasDeadline {
				newCtx, cancel := context.WithDeadline(ctx, deadline)
				defer cancel()
				ctx = newCtx
			}

			r.Handler.ServeHTTP(rec, req.WithContext(ctx))
			return rec.Result(), nil
		}
	}

	// Fallback to default transport for regular requests.
	if r.DefaultTransport != nil {
		return r.DefaultTransport.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
