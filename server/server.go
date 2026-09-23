package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	dochi "github.com/samber/do/http/chi/v2"
	"github.com/samber/do/v2"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/features"
	"github.com/khwong-c/pref-syncs/server/middlewares"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type Authenticator interface {
	Middleware() func(http.Handler) http.Handler
	UserContext(next http.Handler) http.Handler
	IsUser(next http.Handler) http.Handler
	IsAdmin(next http.Handler) http.Handler
	GetUserID(ctx context.Context) uuid.UUID
}

type Server struct {
	*http.Server
	router   chi.Router
	appLogic *features.AppLogic
	auth     Authenticator

	shutdown context.CancelFunc
}

func NewServer(inj do.Injector) (*Server, error) {
	const readHeaderTimeout = 10 * time.Second
	cfg := di.InvokeOrProvide(inj, config.LoadConfig)
	serverCtx, shutdown := context.WithCancel(context.Background())
	r := chi.NewRouter()

	s := &Server{
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
			BaseContext: func(_ net.Listener) context.Context {
				return serverCtx
			},
		},
		router:   r,
		appLogic: di.InvokeOrProvide(inj, features.NewAppLogic),
		auth: di.InvokeOrProvide[Authenticator](inj,
			func(i do.Injector) (Authenticator, error) {
				return middlewares.NewAuthenticator(i, cfg), nil
			},
		),

		shutdown: shutdown,
	}

	r.Use(middleware.StripSlashes)
	r.Use(middleware.NoCache)
	r.Use(middlewares.ErrorBuilder)

	dochi.Use(r, "/debug/di", inj)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})

	if cfg.IDP.Enable {
		const (
			callbackPath = "/auth-cb"
		)

		oidpHandler, err := createLocalIDP(cfg, callbackPath)
		if err != nil {
			return nil, err
		}
		r.Mount(cfg.IDP.Path, oidpHandler)
		r.Get(callbackPath, createLocalIDPCallbackHandler(cfg, callbackPath))
	}

	r.Group(func(r chi.Router) {
		r.Use(s.auth.Middleware())
		r.Use(s.auth.UserContext)
		r.With(s.auth.IsUser).Get("/app/{app}", s.HandleGetApp)
		r.Route("/app", func(r chi.Router) {
			r.Use(s.auth.IsUser)
			r.Use(s.auth.IsAdmin)
			r.Post("/", s.HandleNewApp)
			r.Put("/{app}", s.HandleModifyApp)
			r.Delete("/{app}", s.HandleDeleteApp)
		})

		r.Route("/user", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(s.auth.IsUser)
				r.Get("/", s.HandleGetUser)
				r.Delete("/", s.HandleDeleteUser)
				r.Put("/to/{app}", s.HandleAuthoriseUser)
				r.Delete("/from/{app}", s.HandleDeauthoriseUser)
			})
			// // Capacity to modify other users
			// r.Group(func(r chi.Router) {
			// 	r.Use(s.auth.IsUser)
			// 	r.Use(s.auth.IsAdmin)
			// 	r.Get("/{user}", s.HandleGetUser)
			// 	r.Delete("/{user}", s.HandleDeleteUser)
			// 	r.Put("/{user}/to/{app}", s.HandleAuthoriseUser)
			// 	r.Delete("/{user}/from/{app}", s.HandleDeauthoriseUser)
			// })
		})

		r.Route("/pref", func(r chi.Router) {
			r.Use(s.auth.IsUser)
			r.Post("/{app}", s.HandlePostPref)
			r.Post("/{app}/from/{src}", s.HandlePostPref)
			r.Get("/{app}", s.HandleGetPerf)
			r.Delete("/{app}", s.HandleDeletePref)
			r.Get("/notification/{app}", s.HandleStartNotification)
			r.Get("/notification/{app}/from/{src}", s.HandleStartNotification)
		})
	})

	// r.Route("/auth/{provider}", func(r chi.Router) {
	// 	r.Get("/login", newServer.HandleAuthLogin)
	// 	r.Get("/callback", newServer.HandleAuthCallback)
	// })

	return s, nil
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.shutdown()
	s.Server.Shutdown(ctx)
}
