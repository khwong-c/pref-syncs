package tests

import (
	"net/http"

	"github.com/auth0/go-jwt-middleware/v3/core"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/samber/do/v2"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/server/middlewares"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type MockAuthenticator struct {
	*middlewares.Authenticator
}

func NewAlwaysPassAuthenticator(inj do.Injector) *MockAuthenticator {
	cfg := di.InvokeOrProvide(inj, config.LoadConfig)
	return &MockAuthenticator{
		Authenticator: middlewares.NewAuthenticator(inj, cfg, nil),
	}
}

func (a *MockAuthenticator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := core.SetClaims(r.Context(), &validator.ValidatedClaims{
				RegisteredClaims: validator.RegisteredClaims{
					Issuer:  "Fugu Fish",
					Subject: "Sashimi",
				},
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *MockAuthenticator) IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
