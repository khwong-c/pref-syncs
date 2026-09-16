package server

import (
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/server/middlewares"
)

func (s *Server) MountUserRoutes() {
	r := s.router
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
}

func (s *Server) HandleNewUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO Rewire this block to authware
	uid := uuid.NewV7()
	uidStr := chi.URLParam(r, "user")
	if uidStr != "" {
		if parsedUid, err := uuid.Parse(uidStr); err != nil {
			middlewares.SimpleHTTPError(
				ctx, w, err, "Invalid UUID", http.StatusBadRequest,
			)
			return
		} else {
			uid = parsedUid
		}
	}

	user, err := s.appLogic.CreateUser(ctx, &repos.User{
		ID:           uid,
		AuthProvider: "Temp Provider",
		AuthUserID:   "Temp-User",
	})
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to create user", 0,
		)
		return
	}
	render.JSON(w, r, user)
}

func (s *Server) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO Rewire this block to authware
	userCtx := middlewares.GetUserInfo(ctx)
	oops.FromContext(ctx).Assert(userCtx != nil)
	uid := userCtx.UserID

	uidStr := chi.URLParam(r, "user")
	if uidStr != "" {
		if parsedUid, err := uuid.Parse(uidStr); err != nil {
			middlewares.SimpleHTTPError(
				ctx, w, err, "Invalid UUID", http.StatusBadRequest,
			)
			return
		} else {
			uid = parsedUid
		}
	}

	user, err := s.appLogic.GetUser(ctx, uid)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to get user", 0,
		)
		return
	}
	render.JSON(w, r, user)
}

func (s *Server) HandleAuthoriseUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid, err := uuid.Parse(chi.URLParam(r, "user"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}
	aid, err := uuid.Parse(chi.URLParam(r, "app"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}
	user, err := s.appLogic.AuthoriseUserToApp(ctx, uid, aid)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to authorise user to app", 0,
		)
		return
	}
	render.JSON(w, r, user)
}

func (s *Server) HandleDeauthoriseUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid, err := uuid.Parse(chi.URLParam(r, "user"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}
	aid, err := uuid.Parse(chi.URLParam(r, "app"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}
	user, err := s.appLogic.DeauthoriseUserToApp(ctx, uid, aid)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to deauthorise user to app", 0,
		)
		return
	}
	render.JSON(w, r, user)
}

func (s *Server) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := uuid.MustParse(chi.URLParam(r, "user"))
	if err := s.appLogic.DeleteUser(ctx, uid); err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to delete user", 0,
		)
		return
	}
	render.JSON(w, r, struct {
		Success bool `json:"success"`
	}{true})
}
