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
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

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

type AuthTestSuite struct {
	suite.Suite
	inj    do.Injector
	server *server.Server
	db     *gorm.DB
	cfg    *config.Config

	anonClient  *client.Client
	userClient  *client.Client
	adminClient *client.Client
	httpClient  *http.Client
}

func (s *AuthTestSuite) SetupSuite() {
	ctx := s.T().Context()
	s.inj = do.New()
	s.db = di.InvokeOrProvide(s.inj, sql.NewInMemorySQLite)
	s.cfg = di.InvokeOrProvide(s.inj, config.LoadConfig)
	s.cfg.IDP.Enable = true
	s.cfg.IDP.Path = "/idp"
	s.server = di.InvokeOrProvide(s.inj, server.NewServer)
	s.httpClient = httptestclient.New(s.server.Handler)

	s.anonClient = createAPIClient(s.server)

	t, err := getClientCredentialsToken(
		ctx, s.httpClient,
		models.StubClientPublic,
		models.StubClientSecret,
		[]string{"profile", "email"},
	)
	s.Require().NoError(err)
	s.Require().NotNil(t)
	s.Require().NotEmpty(t.AccessToken)
	s.userClient = createAPIClient(s.server, client.WithAuth(
		&client.BearerAuth{Token: t.AccessToken},
	))

	t, err = getClientCredentialsToken(
		ctx, s.httpClient,
		models.StubClientService,
		models.StubClientSecret,
		[]string{models.StubScope},
	)
	s.Require().NoError(err)
	s.Require().NotNil(t)
	s.Require().NotEmpty(t.AccessToken)
	s.adminClient = createAPIClient(s.server, client.WithAuth(
		&client.BearerAuth{Token: t.AccessToken},
	))
	rsp, err := s.httpClient.Get("http://127.0.0.1:7086/idp/oidc/jwks")
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().Equal(http.StatusOK, rsp.StatusCode)
}

func (s *AuthTestSuite) TestIsUserProtector() {
	tests := []struct {
		name         string
		c            *client.Client
		expectedResp int
	}{
		{
			name:         "anonymous client",
			c:            s.anonClient,
			expectedResp: http.StatusUnauthorized,
		},
		{
			name:         "user client",
			c:            s.userClient,
			expectedResp: http.StatusOK,
		},
		{
			name:         "admin client",
			c:            s.adminClient,
			expectedResp: http.StatusOK,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rsp, err := tc.c.GetUser(s.T().Context())
			if tc.expectedResp != http.StatusOK {
				if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
					s.Equal(tc.expectedResp, err.StatusCode)
				}
			} else {
				s.NotNil(rsp)
			}
		})
	}
}

func (s *AuthTestSuite) TestIsAdminProtector() {
	tests := []struct {
		name         string
		c            *client.Client
		expectedResp int
	}{
		{
			name:         "anonymous client",
			c:            s.anonClient,
			expectedResp: http.StatusUnauthorized,
		},
		{
			name:         "user client",
			c:            s.userClient,
			expectedResp: http.StatusUnauthorized,
		},
		{
			name:         "admin client",
			c:            s.adminClient,
			expectedResp: http.StatusOK,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rsp, err := tc.c.CreateApp(s.T().Context(),
				client.AppUpsertRequest{
					Name: new("AppName"),
					Desc: new("AppDesc"),
				},
			)
			if tc.expectedResp != http.StatusOK {
				if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
					s.Equal(tc.expectedResp, err.StatusCode)
				}
			} else {
				s.NotNil(rsp)
			}
		})
	}
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
