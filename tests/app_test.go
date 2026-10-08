package tests

import (
	"errors"
	"net/http"
	"testing"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/features"
	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tests/httptestclient"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type AppTestSuite struct {
	suite.Suite
	inj      do.Injector
	server   *server.Server
	db       *gorm.DB
	cfg      *config.Config
	dataRepo *repos.DataRepo

	userClient  *client.Client
	adminClient *client.Client
	httpClient  *http.Client
}

func (s *AppTestSuite) SetupSuite() {
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

func (s *AppTestSuite) SetupTest() {
	ctx := s.T().Context()
	_, _ = s.adminClient.DeleteUser(ctx)
	_, _ = s.userClient.DeleteUser(ctx)
}

// Regular users are not allowed to create apps.
func (s *AppTestSuite) TestCreateApp_Unauthorized() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}
	if _, err := s.userClient.CreateApp(ctx, app); s.Error(err) {
		if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
			s.Equal(http.StatusUnauthorized, err.StatusCode)
		}
	}
}

// Admin users can create apps.
func (s *AppTestSuite) TestCreateApp() {
	ctx := s.T().Context()
	// Get Admin User ID
	uid, err := getUserID(ctx, s.adminClient)
	s.Require().NoError(err)

	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}

	// Verify default Apps Values
	rsp, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rsp)

	if s.NotNil(rsp.Name) {
		s.Equal(*app.Name, *rsp.Name)
	}
	if s.NotNil(rsp.Desc) {
		s.Equal(*app.Desc, *rsp.Desc)
	}
	if s.NotNil(rsp.MaxUser) {
		s.Equal(features.DefaultMaxUser, int(*rsp.MaxUser))
	}
	if s.NotNil(rsp.MaxPayload) {
		s.Equal(features.DefaultMaxPayload, int(*rsp.MaxPayload))
	}
	if s.NotNil(rsp.ID) {
		if id, err := uuid.Parse(*rsp.ID); s.NoError(err) {
			s.NotEqual(uuid.Nil(), id)
		}
	}
	if s.NotNil(rsp.CreatedBy) {
		s.Equal(uid, uuid.MustParse(*rsp.CreatedBy))
	}
}

func (s *AppTestSuite) TestGetApp() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}
	rsp, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().NotNil(rsp.ID)
	appID := uuid.MustParse(*rsp.ID)

	if rsp, err = s.adminClient.GetApp(ctx, appID.String()); s.NoError(err) && s.NotNil(rsp) && s.NotNil(rsp.ID) {
		s.Equal(appID, uuid.MustParse(*rsp.ID))
	}
}

func (s *AppTestSuite) TestDeleteApp() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}
	rsp, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().NotNil(rsp.ID)
	appID := uuid.MustParse(*rsp.ID)

	// Ensure App Exists
	rsp, err = s.adminClient.GetApp(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().NotNil(rsp.ID)
	s.Require().Equal(appID, uuid.MustParse(*rsp.ID))

	// Delete the App and ensure it does not exist anymore.
	if rsp, err := s.adminClient.DeleteApp(ctx, appID.String()); s.NoError(err) && s.NotNil(rsp) {
		if s.NotNil(rsp.Success) {
			s.True(*rsp.Success)
		}
	}
	// Check if the App was deleted.
	if _, err := s.adminClient.GetApp(ctx, appID.String()); s.Error(err) {
		if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
			s.Equal(http.StatusNotFound, err.StatusCode)
		}
	}
}

// User can authorise an App to themselves.
func (s *AppTestSuite) TestAppAuthorisation() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}
	userUID, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	// Create an App
	rspCreate, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rspCreate)
	s.Require().NotNil(rspCreate.ID)
	appID := uuid.MustParse(*rspCreate.ID)

	rsp, err := s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	if s.NotNil(rsp.ID) {
		s.Equal(userUID, uuid.MustParse(*rsp.ID))
	}

	// There shall be exactly 1 App authorized to the user.
	if s.Len(rsp.Apps, 1) && s.NotNil(rsp.Apps[0].ID) {
		s.Equal(appID, uuid.MustParse(*rsp.Apps[0].ID))
	}

	// Check for consistency with Get User Call.
	rsp, err = s.userClient.GetUser(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	if s.NotNil(rsp.ID) {
		s.Equal(userUID, uuid.MustParse(*rsp.ID))
	}
	if s.Len(rsp.Apps, 1) && s.NotNil(rsp.Apps[0].ID) {
		s.Equal(appID, uuid.MustParse(*rsp.Apps[0].ID))
	}
}

// User can deauthorise an App from themselves.
func (s *AppTestSuite) TestAppDeauthorise() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}
	userUID, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	// Create an App
	rspCreate, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rspCreate)
	s.Require().NotNil(rspCreate.ID)
	appID := uuid.MustParse(*rspCreate.ID)

	rsp, err := s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)

	// There shall be exactly 1 App authorized to the user.
	if s.Len(rsp.Apps, 1) && s.NotNil(rsp.Apps[0].ID) {
		s.Equal(appID, uuid.MustParse(*rsp.Apps[0].ID))
	}

	// Deauthorise the App from the user. No more apps available to the user.
	rsp, err = s.userClient.DeauthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().Len(rsp.Apps, 0)

	// Check for consistency with Get User Call.
	rsp, err = s.userClient.GetUser(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	if s.NotNil(rsp.ID) {
		s.Equal(userUID, uuid.MustParse(*rsp.ID))
	}
	s.Len(rsp.Apps, 0)
}

// User can authorise an App to themselves multiple times. Yet it shall have no extra effect.
func (s *AppTestSuite) TestAppDuplicatedAuthorisation() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}

	// Create an App
	rspCreate, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rspCreate)
	s.Require().NotNil(rspCreate.ID)
	appID := uuid.MustParse(*rspCreate.ID)

	// Authorise the App to the user for once.
	rsp, err := s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)

	// Authorise the App to the user for twice.
	rsp, err = s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)

	// There shall be exactly 1 App authorized to the user.
	if s.Len(rsp.Apps, 1) && s.NotNil(rsp.Apps[0].ID) {
		s.Equal(appID, uuid.MustParse(*rsp.Apps[0].ID))
	}
}

// All users will be unauthorised from the App when the App is deleted by the creator.
func (s *AppTestSuite) TestUserUnauthorisedByDeletingApp() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}

	// Create an App
	rspCreate, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rspCreate)
	s.Require().NotNil(rspCreate.ID)
	appID := uuid.MustParse(*rspCreate.ID)

	// Authorise the App to the user.
	rsp, err := s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)

	// Delete the App
	rspDel, err := s.adminClient.DeleteApp(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rspDel)
	s.Require().True(*rspDel.Success)

	// There shall be No App authorized to the user.
	if rsp, err := s.userClient.GetUser(ctx); s.NoError(err) && s.NotNil(rsp) {
		s.Len(rsp.Apps, 0)
	}
}

// All users will be unauthorised from the App when the User who created the App is deleted.
func (s *AppTestSuite) TestUserUnauthorisedByDeletingUser() {
	ctx := s.T().Context()
	app := client.AppRequest{
		Name: new("demo-app"),
		Desc: new("Demo Application"),
	}

	// Create an App
	rspCreate, err := s.adminClient.CreateApp(ctx, app)
	s.Require().NoError(err)
	s.Require().NotNil(rspCreate)
	s.Require().NotNil(rspCreate.ID)
	appID := uuid.MustParse(*rspCreate.ID)

	// Authorise the App to the user.
	rsp, err := s.userClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)

	// Delete the Admin User
	rspDel, err := s.adminClient.DeleteUser(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(rspDel)
	s.Require().True(*rspDel.Success)

	// There shall be No App authorized to the user.
	if rsp, err := s.userClient.GetUser(ctx); s.NoError(err) && s.NotNil(rsp) {
		s.Len(rsp.Apps, 0)
	}
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
