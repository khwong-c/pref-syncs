package server

import (
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/server/middlewares"
)

func (s *Server) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid := middlewares.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

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

	uid := middlewares.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

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

	uid := middlewares.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

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

	uid := middlewares.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

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
