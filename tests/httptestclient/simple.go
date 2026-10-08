package httptestclient

import (
	"net/http"
	"net/http/httptest"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// New creates a new http.Client that uses the provided handler to handle requests.
// It acts like a normal http.Client, but no server port is needed and no network calls are made.
func New(handler http.Handler) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			return rec.Result(), nil
		}),
	}
}
