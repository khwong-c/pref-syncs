package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	dochi "github.com/samber/do/http/chi/v2"
	"github.com/samber/do/v2"

	"github.com/khwong-c/pref-syncs/config"
	"github.com/khwong-c/pref-syncs/features"
	"github.com/khwong-c/pref-syncs/server/middlewares"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type Server struct {
	*http.Server
	router   chi.Router
	appLogic *features.AppLogic
	auth     *middlewares.Authenticator

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
		auth:     middlewares.NewAuthenticator(inj, cfg),

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
		r.Route("/app", func(r chi.Router) {
			r.Use(s.auth.IsUser)
			r.Post("/", s.HandleNewApp)
			r.Get("/{id}", s.HandleGetApp)
			r.Put("/{id}", s.HandleGetApp)
			r.Delete("/{id}", s.HandleDeleteApp)
		})

		// TODO: Remove {user} later on by accessing the user from the request context
		r.Route("/user", func(r chi.Router) {
			r.Post("/", s.HandleNewUser)

			r.Group(func(r chi.Router) {
				r.Use(s.auth.IsUser)
				r.Get("/", s.HandleGetUser)
			})
			r.Group(func(r chi.Router) {
				r.Use(s.auth.IsUser)
				r.Use(s.auth.IsAdmin)
				r.Post("/{user}", s.HandleNewUser) // TODO: Temp Routine for testing
				r.Get("/{user}", s.HandleGetUser)  // TODO: Temp Routine for testing
				r.Delete("/{user}", s.HandleDeleteUser)
				r.Put("/{user}/{app}", s.HandleAuthoriseUser)
				r.Delete("/{user}/{app}", s.HandleDeauthoriseUser)
			})
		})

		// TODO: Remove {user} later on by accessing the user from the request context
		r.Route("/pref", func(r chi.Router) {
			r.Use(s.auth.IsUser)
			r.Post("/{user}/{app}", s.HandlePostPref)
			r.Post("/{user}/{app}/from/{src}", s.HandlePostPref)
			r.Get("/{user}/{app}", s.HandleGetPerf)
			r.Delete("/{user}/{app}", s.HandleDeletePref)
			r.Get("/notification/{user}/{app}/from/{src}", s.HandleStartNotification)
			r.Get("/notification/{user}/{app}", s.HandleStartNotification)
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
