package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/samber/oops"
	"github.com/samber/ro"
	"go.jetify.com/sse"

	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/server/middlewares"
)

type prefRspPayload struct {
	UserID    uuid.UUID       `json:"user_id"`
	AppID     uuid.UUID       `json:"app_id"`
	UpdatedAt time.Time       `json:"updated_at"`
	Data      json.RawMessage `json:"data"`
}

func (s *Server) HandlePostPref(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid := s.auth.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

	aid, err := uuid.Parse(chi.URLParam(r, "app"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}

	var srcFilter *string = nil
	if src := chi.URLParam(r, "src"); src != "" {
		srcFilter = &src
	}

	app, err := s.appLogic.GetApp(ctx, aid)
	if err != nil {
		middlewares.HandleHTTPError(ctx, w, err)
		return
	}

	payload, err := s.appLogic.ParsePayloadWithLimit(ctx, r.Body, app)
	defer r.Body.Close()
	if err != nil {
		middlewares.HandleHTTPError(ctx, w, err)
		return
	}
	if !payload.IsValid() {
		err := oops.FromContext(ctx).New("Invalid JSON payload")
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid Payload", http.StatusBadRequest,
		)
		return
	}

	updated, err := s.appLogic.UpdatePref(ctx, &repos.PrefEntry{
		UserID:  uid,
		AppID:   aid,
		Payload: payload.String(),
	}, srcFilter)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to update preference", 0,
		)
		return
	}

	render.JSON(w, r, prefRspPayload{
		UserID:    updated.UserID,
		AppID:     updated.AppID,
		UpdatedAt: updated.UpdatedAt,
		Data:      payload,
	})
}

func (s *Server) HandleGetPerf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid := s.auth.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

	aid, err := uuid.Parse(chi.URLParam(r, "app"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}

	pref, err := s.appLogic.GetPref(ctx, uid, aid)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to retrieve preference", 0,
		)
		return
	}
	rawMsg := json.RawMessage(pref.Payload)
	render.JSON(w, r, prefRspPayload{
		UserID:    uid,
		AppID:     aid,
		UpdatedAt: pref.UpdatedAt,
		Data:      rawMsg,
	})
}

func (s *Server) HandleDeletePref(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid := s.auth.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

	aid, err := uuid.Parse(chi.URLParam(r, "app"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}

	err = s.appLogic.DeletePref(ctx, uid, aid)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to delete preference", 0,
		)
		return
	}
	render.JSON(w, r, struct {
		Success bool `json:"success"`
	}{true})
}

func (s *Server) HandleStartNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	uid := s.auth.GetUserID(ctx)
	oops.FromContext(ctx).Assert(uid != uuid.Nil())

	aid, err := uuid.Parse(chi.URLParam(r, "app"))
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Invalid UUID", http.StatusBadRequest,
		)
		return
	}

	srcStr := chi.URLParam(r, "src")
	var src *string = nil
	if srcStr != "" {
		src = &srcStr
	}

	stream, err := s.appLogic.GetNotificationStream(ctx, uid, aid, src)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to get notification stream", 0,
		)
		return
	}
	_ = stream

	conn, err := sse.Upgrade(
		ctx, w,
		sse.WithCloseMessage(&sse.Event{
			Event: "server-disconnected",
		}),
	)
	if err != nil {
		middlewares.SimpleHTTPError(
			ctx, w, err, "Failed to upgrade connection", http.StatusInternalServerError,
		)
		return
	}
	defer conn.Close()

	observer := ro.OnNextWithContext(
		func(ctx context.Context, value *repos.Notification) {
			if value == nil {
				return
			}
			if err := conn.SendData(ctx, value); err != nil {
				return
			}
			if err := conn.SendEvent(ctx, &sse.Event{
				Event:     "Fugu",
				Data:      value,
				Retry:     0,
				Split:     false,
				Timestamp: time.Time{},
			}); err != nil {
				return
			}
		},
	)
	sub := stream.SubscribeWithContext(ctx, observer)
	defer sub.Unsubscribe()

	// Wait for remote disconnection or server shutdown
	<-ctx.Done()
}
