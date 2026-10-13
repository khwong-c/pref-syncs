package tests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"
	"time"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/tests/httptestclient"
	"github.com/khwong-c/pref-syncs/tests/tags"

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

	userClient        *client.Client
	userStreamClient  *client.Client
	adminClient       *client.Client
	adminStreamClient *client.Client
	httpClient        *http.Client

	testTimeout time.Duration
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
	s.adminStreamClient = createAPIStreamingClient(s.server, client.WithAuth(
		&client.BearerAuth{Token: t.AccessToken},
	))

	s.testTimeout = 3 * time.Second
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

func (s *NotificationTestSuite) TestNotifiedByPostPref() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()
	uid, err := getUserID(ctx, s.userClient)
	s.Require().NoError(err)

	const (
		source  = "random-device"
		content = "Random Content"
	)

	tc := []struct {
		name     string
		postFunc func(ctx context.Context) error
		src      *string
	}{
		{
			name: "Without Source",
			postFunc: func(ctx context.Context) error {
				_, err := s.userClient.PostPref(ctx, appID.String(), content)
				return err
			},
			src: nil,
		},
		{
			name: "With Source",
			postFunc: func(ctx context.Context) error {
				_, err := s.userClient.PostPrefFromSource(
					ctx, appID.String(), source, content,
				)
				return err
			},
			src: new(source),
		},
	}
	for _, t := range tc {
		s.Run(t.name, func() {
			ctx, cancel := context.WithTimeout(s.T().Context(), s.testTimeout)
			defer cancel()
			notifications, err := s.userStreamClient.StartNotificationStream(ctx, appID.String())
			s.Require().NoError(err)
			s.Require().NotNil(notifications)
			postErr := t.postFunc(ctx)
			s.Require().NoError(postErr)

			resp, err := notifications.Next()
			if s.NoError(err) && s.NotNil(resp) {
				if s.NotNil(resp.User) {
					s.Equal(uid, uuid.MustParse(*resp.User))
				}
				if s.NotNil(resp.App) {
					s.Equal(appID, uuid.MustParse(*resp.App))
				}
				if t.src == nil {
					s.Nil(resp.Src)
				}
				if t.src != nil && s.NotNil(resp.Src) {
					s.Equal(*t.src, *resp.Src)
				}
			}
		})
	}
}

func (s *NotificationTestSuite) TestNotificationFilteredBySource() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	const content = "Random Data"
	const src1 = "src1"
	const src2 = "src2"

	postPref := func(ctx context.Context, _ string) error {
		_, err := s.userClient.PostPref(ctx, appID.String(), content)
		return err
	}
	postPrefFromSrc := func(ctx context.Context, src string) error {
		_, err := s.userClient.PostPrefFromSource(ctx, appID.String(), src, content)
		return err
	}
	subAll := func(ctx context.Context, _ string) (*client.EventStream[client.Notification], error) {
		return s.userStreamClient.StartNotificationStream(ctx, appID.String())
	}
	subFromSrc := func(ctx context.Context, src string) (*client.EventStream[client.Notification], error) {
		return s.userStreamClient.StartNotificationFromSourceStream(ctx, appID.String(), src)
	}

	postFrom := []struct {
		source   string
		postFunc func(ctx context.Context, src string) error
	}{
		{"", postPref},
		{src1, postPrefFromSrc},
		{src2, postPrefFromSrc},
	}
	type subEntry struct {
		name     string
		source   string
		subFunc  func(ctx context.Context, src string) (*client.EventStream[client.Notification], error)
		expected []*string
	}
	subFrom := []subEntry{
		{"all", "", subAll, []*string{nil, new(src1), new(src2)}},
		{src1, src1, subFromSrc, []*string{nil, new(src2)}},
		{src2, src2, subFromSrc, []*string{nil, new(src1)}},
	}

	ctxStrm, cancel := context.WithCancel(ctx)
	streams := lo.Map(subFrom, func(item subEntry, i int) *client.EventStream[client.Notification] {
		strm, err := item.subFunc(ctxStrm, item.source)
		s.Require().NoError(err)
		return strm
	})

	for _, req := range postFrom {
		s.Require().NoError(
			req.postFunc(ctx, req.source),
		)
	}
	cancel()

	results := make([][]*string, len(streams))
	for i, strm := range streams {
		results[i] = make([]*string, 0, len(postFrom))
		for {
			item, err := strm.Next()
			if err != nil {
				s.Equal(err, io.EOF)
				break
			}
			if s.NotNil(item) {
				results[i] = append(results[i], item.Src)
			}
		}
		strm.Close()
	}

	for ti, t := range subFrom {
		s.Run(t.name, func() {
			s.Require().Len(results[ti], len(t.expected))
			for i, expected := range t.expected {
				if expected == nil {
					s.Nil(results[ti][i])
					continue
				}
				if s.NotNil(results[ti][i]) {
					s.Equal(*expected, *results[ti][i])
				}
			}
		})
	}
}

func (s *NotificationTestSuite) TestNotificationIsolatedByUser() {
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()
	_, err := s.adminClient.AuthoriseUser(ctx, appID.String())
	s.Require().NoError(err)

	const payloadStub = "data-payload"
	const userNotifications = 1
	const adminNotifications = 3

	ctxStrm, cancel := context.WithCancel(ctx)
	userStrm, err := s.userStreamClient.StartNotificationStream(ctxStrm, appID.String())
	s.Require().NoError(err)
	adminStrm, err := s.adminStreamClient.StartNotificationStream(ctxStrm, appID.String())
	s.Require().NoError(err)

	for range userNotifications {
		_, err = s.userClient.PostPref(ctx, appID.String(), payloadStub)
		s.Require().NoError(err)
	}
	for range adminNotifications {
		_, err = s.adminClient.PostPref(ctx, appID.String(), payloadStub)
		s.Require().NoError(err)
	}
	cancel()

	userResult := make([]string, 0)
	adminResult := make([]string, 0)
	for {
		item, err := userStrm.Next()
		if err != nil {
			s.Equal(err, io.EOF)
			break
		}
		if s.NotNil(item) && s.NotNil(item.User) {
			userResult = append(userResult, *item.User)
		}
	}
	for {
		item, err := adminStrm.Next()
		if err != nil {
			s.Equal(err, io.EOF)
			break
		}
		if s.NotNil(item) && s.NotNil(item.User) {
			adminResult = append(adminResult, *item.User)
		}
	}
	userStrm.Close()
	adminStrm.Close()

	s.Len(userResult, userNotifications)
	s.Len(adminResult, adminNotifications)
	s.NotEqual(userResult, adminResult)
}

func (s *NotificationTestSuite) TestNotificationIsolatedByApp() {
	const (
		payloadStub       = "data-payload"
		app1Notifications = 1
		app2Notifications = 3
	)

	ctx := s.T().Context()
	appID1, appID2 := s.createAndAuthoriseApp(), s.createAndAuthoriseApp()

	ctxStrm, cancel := context.WithCancel(ctx)
	strm1, err := s.userStreamClient.StartNotificationStream(ctxStrm, appID1.String())
	s.Require().NoError(err)
	strm2, err := s.userStreamClient.StartNotificationStream(ctxStrm, appID2.String())
	s.Require().NoError(err)

	for range app1Notifications {
		_, err = s.userClient.PostPref(ctx, appID1.String(), payloadStub)
		s.Require().NoError(err)
	}
	for range app2Notifications {
		_, err = s.userClient.PostPref(ctx, appID2.String(), payloadStub)
		s.Require().NoError(err)
	}
	cancel()

	app1Result, app2Result := make([]string, 0), make([]string, 0)
	for {
		item, err := strm1.Next()
		if err != nil {
			s.Equal(err, io.EOF)
			break
		}
		if s.NotNil(item) && s.NotNil(item.App) {
			app1Result = append(app1Result, *item.App)
		}
	}
	for {
		item, err := strm2.Next()
		if err != nil {
			s.Equal(err, io.EOF)
			break
		}
		if s.NotNil(item) && s.NotNil(item.App) {
			app2Result = append(app2Result, *item.App)
		}
	}
	strm1.Close()
	strm2.Close()

	s.Len(app1Result, app1Notifications)
	s.Len(app2Result, app2Notifications)
	s.NotEqual(app1Result, app2Result)
	s.Equal(appID1.String(), app1Result[0])
	s.Equal(appID2.String(), app2Result[0])
}

func (s *NotificationTestSuite) TestSubscribeNonExistsApp() {
	appID := uuid.New()
	ctx := s.T().Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err := s.userStreamClient.StartNotificationStream(ctx, appID.String())
	if s.Error(err) {
		if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
			s.Equal(http.StatusNotFound, err.StatusCode)
		}
	}
}

func (s *NotificationTestSuite) TestSubscribeUnauthorisedApp() {
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

	// Skip Authorisation. Subscribe notification stream from the app.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err = s.userStreamClient.StartNotificationStream(ctx, appID.String())
	if s.Error(err) {
		if err, ok := errors.AsType[*client.APIError](err); s.True(ok) && s.NotNil(err) {
			s.Equal(http.StatusUnauthorized, err.StatusCode)
		}
	}
}

// Detect potential memory leaks from the notification stream.
func (s *NotificationTestSuite) TestMemoryUsage() {
	if !tags.Tags.LongTest {
		s.T().Skip("skipping memory test without long tag")
	}
	const memoryTolerance = 0.1
	ctx := s.T().Context()
	appID := s.createAndAuthoriseApp()

	mem := make([]runtime.MemStats, 3)

	// Run the load test for multiple times.
	// The first run acts as the baseline.
	// Memory usage after Garbage collection of the second run will be higher significantly
	// if the notification streams handler is causing memory leaks.
	for i := range mem {
		// Free memory before and after each run.
		// Collecting garbage for multiple iterations to ensure that the memory usage is stable.
		for range 5 {
			runtime.GC()
		}
		// Creating notification streams.
		for range 1_000 {
			// Allow the client to disconnect after each stream is tested.
			ctxStrm, cancel := context.WithCancel(ctx)
			notifications, err := s.userStreamClient.StartNotificationStream(ctxStrm, appID.String())
			s.Require().NoError(err)
			s.Require().NotNil(notifications)

			// Triggering notifications for a few times.
			for j := range 5 {
				rsp, err := s.userClient.PostPref(ctx, appID.String(), fmt.Sprintf("fugu %d", j))
				s.Require().NoError(err)
				s.Require().NotNil(rsp)
				// Consume the notification.
				rsp2, err := notifications.Next()
				s.Require().NoError(err)
				s.Require().NotNil(rsp2)
			}

			// Close the notification stream.
			_ = notifications.Close()
			cancel()
		}
		// Freeing the memory
		for range 5 {
			runtime.GC()
		}
		runtime.ReadMemStats(&mem[i])
		// Check if the memory usage is within tolerance.
		if i > 0 {
			s.LessOrEqual(float64(mem[i].HeapInuse), float64(mem[0].HeapInuse)*(1+memoryTolerance))
		}
	}
}
