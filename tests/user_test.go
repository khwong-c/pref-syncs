package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tests/client"
	"github.com/khwong-c/pref-syncs/tests/httptestclient"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type UserTestSuite struct {
	suite.Suite
	inj      do.Injector
	server   *server.Server
	db       *gorm.DB
	cfg      *config.Config
	dataRepo *repos.DataRepo

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
}

func getUserID(ctx context.Context, client *client.Client) (uuid.UUID, error) {
	rsp, err := client.GetUser(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	if rsp.ID == nil {
		return uuid.UUID{}, fmt.Errorf("user ID is nil")
	}
	return uuid.Parse(*rsp.ID)
}

func (s *UserTestSuite) SetupTest() {
	_, _ = s.userClient.DeleteUser(s.T().Context())
}

func (s *UserTestSuite) TestCreateUserByGetUser() {
	ctx := s.T().Context()
	if rsp, err := s.userClient.GetUser(ctx); s.NoError(err) {
		if s.NotNil(rsp.AuthProvider) {
			s.Equal(fmt.Sprintf("http://127.0.0.1:%d%s", s.cfg.Port, s.cfg.IDP.Path), *rsp.AuthProvider)
		}
		if s.NotNil(rsp.AuthSub) {
			s.Equal(models.StubClientPublic, *rsp.AuthSub)
		}
		if s.NotNil(rsp.ID) {
			_, err := uuid.Parse(*rsp.ID)
			s.NoError(err)
		}
	}
}

func (s *UserTestSuite) TestGetExistingUser() {
	ctx := s.T().Context()

	// Create a user by Get User Call
	uid, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	// Get user again. The ID shall match.
	if rsp, err := s.userClient.GetUser(ctx); s.NoError(err) && s.NotNil(rsp.ID) {
		s.Equal(uid, uuid.MustParse(*rsp.ID))
	}

	if user, err := s.dataRepo.GetUser(ctx, uid); s.NoError(err) && s.NotNil(user) {
		s.Equal(uid, user.ID)
	}
}

func (s *UserTestSuite) TestUserIDNotNull() {
	ctx := s.T().Context()
	if rsp, err := s.userClient.GetUser(ctx); s.NoError(err) && s.NotNil(rsp.ID) {
		if uid, err := uuid.Parse(*rsp.ID); s.NoError(err) {
			s.False(uid == uuid.Nil(), "user id should not be nil")
		}
	}
}

func (s *UserTestSuite) TestDeleteUser() {
	ctx := s.T().Context()
	// Create a user by Get User Call
	uid, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	// Delete the user
	if rsp, err := s.userClient.DeleteUser(ctx); s.NoError(err) && s.NotNil(rsp.Success) {
		s.True(*rsp.Success)
	}

	// Check if the user is removed from the data record.
	if _, err := s.dataRepo.GetUser(ctx, uid); s.Error(err) {
		s.True(models.IsNotFoundErr(err))
	}

	// Get user again. The ID shall not match.
	// Create a user by Get User Call
	if uid2, err := getUserID(ctx, s.userClient); s.NoError(err) {
		s.NotEqual(uid, uid2)
	}
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
