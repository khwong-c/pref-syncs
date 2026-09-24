package tests

import (
	"errors"
	"net/http"
	"testing"

	"github.com/khwong-c/httptestclient"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type SmokeTestSuite struct {
	suite.Suite
}

func (s *SmokeTestSuite) TestSmoke() {
	inj := do.New()
	_ = di.InvokeOrProvide(inj, sql.NewInMemorySQLite)
	svr := di.InvokeOrProvide(inj, server.NewServer)
	c := client.NewClient(
		"http://localhost:7086",
		client.WithHTTPClient(
			httptestclient.New(svr.Handler),
		),
	)
	if rsp, err := c.GetRoot(s.T().Context()); s.NoError(err) && s.NotNil(rsp) {
		s.Equal("Hello World", *rsp)
	}
}

func (s *SmokeTestSuite) TestIDPEndpoints() {

	tests := []struct {
		name     string
		enable   bool
		oidcResp int
		cbResp   int
	}{
		{
			name:     "disabled",
			enable:   false,
			oidcResp: http.StatusNotFound,
			cbResp:   http.StatusNotFound,
		},
		{
			name:     "enabled",
			enable:   true,
			oidcResp: http.StatusOK,
			cbResp:   http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			inj := do.New()
			_ = di.InvokeOrProvide(inj, sql.NewInMemorySQLite)
			cfg := di.InvokeOrProvide(inj, config.LoadConfig)
			cfg.IDP.Enable = tc.enable
			svr := di.InvokeOrProvide(inj, server.NewServer)

			httpClient := httptestclient.New(svr.Handler)
			c := createAPIClient(svr)

			// Check if Auth Callback exists
			_, err := c.GetAuthCallback(s.T().Context(), client.GetAuthCallbackParams{})
			if err, ok := errors.AsType[*client.APIError](err); s.True(ok) {
				s.Equal(tc.cbResp, err.StatusCode)
			}

			// Check if OIDC Well-Known Host Introspection Endpoint exists
			if rsp, err := httpClient.Get(
				"/idp/.well-known/openid-configuration",
			); s.NoError(err) {
				s.Equal(tc.oidcResp, rsp.StatusCode)
			}
		})
	}
}

func (s *SmokeTestSuite) TestOAuth2ClientCredentials() {
	inj := do.New()
	_ = di.InvokeOrProvide(inj, sql.NewInMemorySQLite)
	cfg := di.InvokeOrProvide(inj, config.LoadConfig)
	cfg.IDP.Enable = true
	cfg.IDP.Path = "/idp"
	svr := di.InvokeOrProvide(inj, server.NewServer)

	httpClient := httptestclient.New(svr.Handler)

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		scopes       []string
	}{
		{
			name:         "public-rp-client",
			clientID:     models.StubClientPublic,
			clientSecret: models.StubClientSecret,
			scopes:       []string{"profile", "email"},
		},
		{
			name:         "service-client",
			clientID:     models.StubClientService,
			clientSecret: models.StubClientSecret,
			scopes:       []string{models.StubScope},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if token, err := getClientCredentialsToken(
				s.T().Context(),
				httpClient,
				tc.clientID,
				tc.clientSecret,
				tc.scopes,
			); s.NoError(err) && s.NotNil(token) {
				s.NotEmpty(token.AccessToken)
			}
		})
	}
}

func TestSmokeTestSuite(t *testing.T) {
	// Run test suite
	suite.Run(t, new(SmokeTestSuite))
}
