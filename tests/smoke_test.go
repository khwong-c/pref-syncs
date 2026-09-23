package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/khwong-c/httptestclient"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type SmokeTestSuite struct {
	suite.Suite
	inj    do.Injector
	server *server.Server
	db     *gorm.DB
	cfg    *config.Config
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
	if rsp, err := c.GetRoot(context.Background()); s.NoError(err) && s.NotNil(rsp) {
		s.Equal(*rsp, "Hello World")
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

			c, httpClient := createClients(svr)

			// Check if Auth Callback exists
			_, err := c.GetAuthCallback(context.Background(), client.GetAuthCallbackParams{})
			if err, ok := errors.AsType[*client.APIError](err); s.True(ok) {
				s.Equal(err.StatusCode, tc.cbResp)
			}

			// Check if OIDC Well-Known Host Introspection Endpoint exists
			if rsp, err := httpClient.Get(
				"/idp/.well-known/openid-configuration",
			); s.NoError(err) {
				s.Equal(rsp.StatusCode, tc.oidcResp)
			}
		})
	}
}

func getClientCredentialsToken(
	ctx context.Context,
	httpClient *http.Client,
	clientID string,
	clientSecret string,
	scopes []string,
) (*oauth2.Token, error) {
	oauthConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "/idp/oidc/token",
		Scopes:       scopes,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	return oauthConfig.Token(ctx)
}

func (s *SmokeTestSuite) TestOAuth2ClientCredentials() {
	inj := do.New()
	_ = di.InvokeOrProvide(inj, sql.NewInMemorySQLite)
	cfg := di.InvokeOrProvide(inj, config.LoadConfig)
	cfg.IDP.Enable = true
	svr := di.InvokeOrProvide(inj, server.NewServer)

	_, httpClient := createClients(svr)

	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		scopes       []string
	}{
		{
			name:         "public-rp-client",
			clientID:     server.StubClientPublic,
			clientSecret: server.StubClientSecret,
			scopes:       []string{"profile", "email"},
		},
		{
			name:         "service-client",
			clientID:     server.StubClientService,
			clientSecret: server.StubClientSecret,
			scopes:       []string{server.StubScope},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if token, err := getClientCredentialsToken(
				context.Background(),
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
