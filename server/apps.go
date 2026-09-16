package server

import (
	"encoding/json/v2"
	"net/http"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/server/middlewares"
)

func (s *Server) MountAppRoutes() {
	r := s.router
	r.Route("/app", func(r chi.Router) {
		r.Post("/", s.HandleNewApp)
		r.Get("/{id}", s.HandleGetApp)
		r.Put("/{id}", s.HandleGetApp)
		r.Delete("/{id}", s.HandleDeleteApp)
	})
}

type appReqPayload struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type appRspPayload struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Desc string    `json:"desc"`
}

func (s *Server) HandleNewApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	payload := &appReqPayload{}
	if err := json.UnmarshalRead(r.Body, payload); err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid Payload", http.StatusBadRequest,
		)
		return
	}
	newApp, err := s.appLogic.CreateApp(ctx, &repos.App{
		Name: payload.Name,
		Desc: payload.Desc,
	})
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to create App", 0,
		)
		return
	}
	render.JSON(w, r, appRspPayload{
		ID:   newApp.ID,
		Name: newApp.Name,
		Desc: newApp.Desc,
	})
}

func (s *Server) HandleGetApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}
	app, err := s.appLogic.GetApp(ctx, id)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to retrieve app", 0,
		)
		return
	}
	render.JSON(w, r, appRspPayload{
		ID:   app.ID,
		Name: app.Name,
		Desc: app.Desc,
	})
}

func (s *Server) HandleModifyApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}

	payload := &appReqPayload{}
	if err := json.UnmarshalRead(r.Body, payload); err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid Payload", http.StatusBadRequest,
		)
		return
	}

	modApp, err := s.appLogic.ModifyApp(ctx, id, &repos.App{
		Name: payload.Name,
		Desc: payload.Desc,
	})
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to modify App", 0,
		)
		return
	}
	render.JSON(w, r, appRspPayload{
		ID:   modApp.ID,
		Name: modApp.Name,
		Desc: modApp.Desc,
	})
}

func (s *Server) HandleDeleteApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}
	err = s.appLogic.DeleteApp(ctx, id)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to delete app", 0,
		)
		return
	}
	render.JSON(w, r, struct {
		Success bool `json:"success"`
	}{true})
}
