package tests

import (
	"context"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/khwong-c/httptestclient"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type NotificationTestSuite struct {
	suite.Suite
	inj      do.Injector
	server   *server.Server
	db       *gorm.DB
	cfg      *config.Config
	dataRepo *repos.DataRepo

	userClient       *client.Client
	userStreamClient *client.Client
	adminClient      *client.Client
	httpClient       *http.Client
}

func (s *NotificationTestSuite) SetupSuite() {
	ctx := s.T().Context()
	s.inj = do.New()
	s.db = di.InvokeOrProvide(s.inj, sql.NewInMemorySQLite)
	s.cfg = di.InvokeOrProvide(s.inj, config.LoadConfig)
	s.cfg.IDP.Enable = true
	s.cfg.IDP.Path = "/idp"
	s.server = di.InvokeOrProvide(s.inj, server.NewServer)
	s.dataRepo = di.Invoke[*repos.DataRepo](s.inj)
	s.httpClient = httptestclient.New(s.server.Handler)

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
	s.userStreamClient = createAPIStreamingClient(s.server, client.WithAuth(
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
}

func (s *NotificationTestSuite) SetupTest() {
	ctx := s.T().Context()
	_, _ = s.adminClient.DeleteUser(ctx)
	_, _ = s.userClient.DeleteUser(ctx)
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}

func (s *NotificationTestSuite) createAndAuthoriseApp() uuid.UUID {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}

	// Create an App
	rsp, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().NotNil(rsp.ID)
	appID := uuid.MustParse(*rsp.ID)

	_, err = s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)

	return appID
}
func (s *NotificationTestSuite) TestGetNotificationStream() {
	ctx := s.T().Context()

	appID := s.createAndAuthoriseApp()

	ctxStream, cancel := context.WithCancel(ctx)
	//defer cancel()
	notifications, err := s.userStreamClient.StartNotificationStream(ctxStream, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(notifications)
	time.Sleep(1 * time.Second)
	postResp, postErr := s.userClient.PostPref(ctx, appID.String(), "fugu")
	s.T().Logf("postResp: %+v, postErr: %v", postResp, postErr)
	time.Sleep(1 * time.Second)
	resp, err := notifications.Next()
	s.T().Log(resp)
	s.T().Log(err)
	cancel()
}
