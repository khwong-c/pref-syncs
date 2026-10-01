package tests

import (
	"net/http"
	"testing"

	"github.com/khwong-c/httptestclient"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type UserTestSuite struct {
	suite.Suite
	inj    do.Injector
	server *server.Server
	db     *gorm.DB
	cfg    *config.Config

	userClient *client.Client
	httpClient *http.Client
}

func (s *UserTestSuite) SetupSuite() {
	ctx := s.T().Context()
	s.inj = do.New()
	s.db = di.InvokeOrProvide(s.inj, sql.NewInMemorySQLite)
	s.cfg = di.InvokeOrProvide(s.inj, config.LoadConfig)
	s.cfg.IDP.Enable = true
	s.cfg.IDP.Path = "/idp"
	s.server = di.InvokeOrProvide(s.inj, server.NewServer)
	s.httpClient = httptestclient.New(s.server.Handler)

	t, err := getClientCredentialsToken(
		ctx, s.httpClient,
		models.StubClientService,
		models.StubClientSecret,
		[]string{"profile", "email"},
	)
	s.Require().NoError(err)
	s.Require().NotNil(t)
	s.Require().NotEmpty(t.AccessToken)
	s.userClient = createAPIClient(s.server, client.WithAuth(
		&client.BearerAuth{Token: t.AccessToken},
	))
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
