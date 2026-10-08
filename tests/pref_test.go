package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/samber/lo"
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

type PrefTestSuite struct {
	suite.Suite
	inj      do.Injector
	server   *server.Server
	db       *gorm.DB
	cfg      *config.Config
	dataRepo *repos.DataRepo

	userClient  *client.Client
	adminClient *client.Client
	httpClient  *http.Client

	timeTolerance time.Duration
}

func (s *PrefTestSuite) SetupSuite() {
	ctx := s.T().Context()
	s.timeTolerance = 5 * time.Millisecond
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

func (s *PrefTestSuite) SetupTest() {
	ctx := s.T().Context()
	_, _ = s.adminClient.DeleteUser(ctx)
	_, _ = s.userClient.DeleteUser(ctx)
}

func TestPrefTestSuite(t *testing.T) {
	suite.Run(t, new(PrefTestSuite))
}

func (s *PrefTestSuite) createAndAuthoriseApp() uuid.UUID {
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

func (s *PrefTestSuite) TestPostPref() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	uid, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	// Posting Preference with or without source specified shall behave the same in this test
	tc := []struct {
		name       string
		postMethod func(ctx context.Context, payload map[string]any) (*client.PrefResponse, error)
	}{
		{
			name: "post pref",
			postMethod: func(ctx context.Context, payload map[string]any) (*client.PrefResponse, error) {
				return s.userClient.PostPref(ctx, appID.String(), payload)
			},
		},
		{
			name: "post pref with source specified",
			postMethod: func(ctx context.Context, payload map[string]any) (*client.PrefResponse, error) {
				return s.userClient.PostPrefFromSource(ctx, appID.String(), "random-source", payload)
			},
		},
	}
	for _, t := range tc {
		s.Run(t.name, func() {
			payload := map[string]any{
				"test_name": t.name,
				"key1":      "value1",
				"key2":      "value2",
			}
			if rsp, err := t.postMethod(ctx, payload); s.NoError(err) && s.NotNil(rsp) {
				if s.NotNil(rsp.UserID) {
					s.Equal(uid.String(), *rsp.UserID)
				}
				if s.NotNil(rsp.AppID) {
					s.Equal(appID.String(), *rsp.AppID)
				}
				if s.NotNil(rsp.UpdatedAt) {
					s.WithinDuration(time.Now(), *rsp.UpdatedAt, s.timeTolerance)
				}
				s.Equal(payload, rsp.Data)
			}
		})
	}
}

func (s *PrefTestSuite) TestGetPref() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	uid, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	// Getting Preference with or without source specified shall behave the same in this test
	tc := []struct {
		name       string
		postMethod func(ctx context.Context, payload map[string]any) (*client.PrefResponse, error)
	}{
		{
			name: "post pref",
			postMethod: func(ctx context.Context, payload map[string]any) (*client.PrefResponse, error) {
				return s.userClient.PostPref(ctx, appID.String(), payload)
			},
		},
		{
			name: "post pref with source specified",
			postMethod: func(ctx context.Context, payload map[string]any) (*client.PrefResponse, error) {
				return s.userClient.PostPrefFromSource(ctx, appID.String(), "random-source", payload)
			},
		},
	}
	for _, t := range tc {
		s.Run(t.name, func() {
			payload := map[string]any{
				"test_name": t.name,
				"key1":      "value1",
				"key2":      "value2",
			}
			rspPost, err := t.postMethod(ctx, payload)
			s.Require().NoError(err)
			s.Require().NotNil(rspPost)

			rsp, err := s.userClient.GetPref(ctx, appID.String())
			s.Require().NoError(err)
			s.Require().NotNil(rsp)

			if s.NotNil(rsp.UserID) {
				s.Equal(uid.String(), *rsp.UserID)
			}
			if s.NotNil(rsp.AppID) {
				s.Equal(appID.String(), *rsp.AppID)
			}
			if s.NotNil(rsp.UpdatedAt) {
				s.WithinDuration(time.Now(), *rsp.UpdatedAt, s.timeTolerance)
			}
			s.Equal(payload, rsp.Data)
		})
	}
}

// The Post Pref API allows users to upsert preferences. Getting the preferences will return the latest preferences submitted.
func (s *PrefTestSuite) TestUpsertPref() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	payload := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	rspPost, err := s.userClient.PostPref(ctx, appID.String(), payload)
	s.Require().NoError(err)
	s.Require().NotNil(rspPost)
	s.Require().NotNil(rspPost.UpdatedAt)
	updateTime1 := *rspPost.UpdatedAt

	newPayload := map[string]any{
		"key1": "new_value1",
		"key2": "new_value2",
	}
	rspPost, err = s.userClient.PostPref(ctx, appID.String(), newPayload)
	s.Require().NoError(err)
	s.Require().NotNil(rspPost)
	s.Require().NotNil(rspPost.UpdatedAt)
	updateTime2 := *rspPost.UpdatedAt

	s.Require().True(updateTime2.After(updateTime1))

	rsp, err := s.userClient.GetPref(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Equal(newPayload, rsp.Data)
	s.NotEqual(payload, rsp.Data)
}

// When a user associates with an app without posting any preference,
// the server returns HTTP 404 if the user tries to retrieve the preferences.
func (s *PrefTestSuite) TestNewAppReturnNoPref() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	if _, err := s.userClient.GetPref(ctx, appID.String()); s.Error(err) {
		if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
			s.Equal(http.StatusNotFound, err.StatusCode)
		}
	}
}

// A set of scenarios that trigger the removal of preferences.
func (s *PrefTestSuite) TestChainPrefRemoval() {
	tc := []struct {
		name       string
		removeFunc func(ctx context.Context, appID string) error
	}{
		{
			name: "deauthorise", // User deauthorise him/herself from the app
			removeFunc: func(ctx context.Context, appID string) error {
				_, err := s.userClient.DeauthoriseUser(ctx, appID)
				return err
			},
		},
		{
			name: "delete pref", // User deletes his/her preference
			removeFunc: func(ctx context.Context, appID string) error {
				_, err := s.userClient.DeletePref(ctx, appID)
				return err
			},
		},
		{
			name: "delete app", // App creator deletes the app
			removeFunc: func(ctx context.Context, appID string) error {
				_, err := s.adminClient.DeleteApp(ctx, appID)
				return err
			},
		},
		{
			name: "delete user", // User deletes him/herself
			removeFunc: func(ctx context.Context, appID string) error {
				_, err := s.userClient.DeleteUser(ctx)
				return err
			},
		},
		{
			name: "delete creator user", // App Creator deletes his/her user
			removeFunc: func(ctx context.Context, appID string) error {
				_, err := s.adminClient.DeleteUser(ctx)
				return err
			},
		},
	}

	for _, t := range tc {
		s.Run(t.name, func() {
			ctx := s.T().Context()
			appID := s.createAndAuthoriseApp()
			uid, err := getUserID(ctx, s.userClient)
			s.Require().NoError(err)

			payload := map[string]any{
				"test_name": t.name,
				"key1":      "value1",
				"key2":      "value2",
			}

			rsp, err := s.userClient.PostPref(ctx, appID.String(), payload)
			s.Require().NoError(err)
			s.Require().NotNil(rsp)

			rsp, err = s.userClient.GetPref(ctx, appID.String())
			s.Require().NoError(err)
			s.Require().NotNil(rsp)
			s.Require().Equal(payload, rsp.Data)

			// Triggering the removing scenario
			s.Require().NoError(
				t.removeFunc(ctx, appID.String()),
			)

			// Check that the preference is removed
			if _, err = s.userClient.GetPref(ctx, appID.String()); s.Error(err) {
				if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
					s.Equal(http.StatusNotFound, err.StatusCode)
				}
			}

			// Confirm this from the repository layer as well.
			if _, err = s.dataRepo.GetPreference(ctx, uid, appID); s.Error(err) {
				s.True(models.IsNotFoundErr(err))
			}
		})
	}
}

func (s *PrefTestSuite) TestSizeLimit() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	rsp, err := s.userClient.GetApp(ctx, appID.String())
	s.Require().NoError(err)
	s.Require().NotNil(rsp)
	s.Require().NotNil(rsp.MaxPayload)
	maxSize := int(*rsp.MaxPayload)

	// Getting Preference with or without source specified shall behave the same in this test
	tc := []struct {
		name       string
		postMethod func(ctx context.Context, payload any) (*client.PrefResponse, error)
	}{
		{
			name: "post pref",
			postMethod: func(ctx context.Context, payload any) (*client.PrefResponse, error) {
				return s.userClient.PostPref(ctx, appID.String(), payload)
			},
		},
		{
			name: "post pref with source specified",
			postMethod: func(ctx context.Context, payload any) (*client.PrefResponse, error) {
				return s.userClient.PostPrefFromSource(ctx, appID.String(), "random-source", payload)
			},
		},
	}
	for _, t := range tc {
		s.Run(t.name, func() {
			subTc := []struct {
				size    int
				success bool
			}{
				{maxSize - 1, true},
				{maxSize, true},
				{maxSize + 1, false},
			}
			for _, subT := range subTc {
				s.Run(fmt.Sprint(subT.size), func() {
					const firstPayload = ""
					_, err := s.userClient.PostPref(ctx, appID.String(), firstPayload)
					s.Require().NoError(err)

					// Create a string payload with specific size
					payload := lo.RandomString(subT.size-2, []rune{'a'})
					rsp, err := t.postMethod(ctx, payload)
					if subT.success {
						s.Require().NoError(err)
						s.Require().NotNil(rsp)
						s.Require().NotNil(rsp.Data)
						s.Equal(payload, rsp.Data)
					} else {
						s.Require().Error(err)
						if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
							s.Equal(http.StatusBadRequest, err.StatusCode)
						}
					}
				})
			}
		})
	}
}
