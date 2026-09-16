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

	newServer := &Server{
		Server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
			BaseContext: func(_ net.Listener) context.Context {
				return serverCtx
			},
		},
		router:   r,
		appLogic: features.NewAppLogic(inj),
		auth:     middlewares.NewAuthenticator(inj),

		shutdown: shutdown,
	}

	r.Use(middleware.StripSlashes)
	r.Use(middleware.NoCache)
	r.Use(middlewares.ErrorBuilder)
	r.Use(newServer.auth.UserContext)

	dochi.Use(r, "/debug/di", inj)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})
	newServer.MountAppRoutes()
	newServer.MountPrefRoutes()
	newServer.MountUserRoutes()

	if cfg.IDP.Enable {
		const oidcPath = "/idp"
		oidpHandler, err := createLocalIDP(cfg, oidcPath)
		if err != nil {
			return nil, err
		}
		r.Mount(oidcPath, oidpHandler)
	}

	// r.Route("/auth/{provider}", func(r chi.Router) {
	// 	r.Get("/login", newServer.HandleAuthLogin)
	// 	r.Get("/callback", newServer.HandleAuthCallback)
	// })

	return newServer, nil
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.shutdown()
	s.Server.Shutdown(ctx)
}
