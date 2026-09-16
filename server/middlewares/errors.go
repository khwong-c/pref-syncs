package middlewares

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/dotse/slug"
	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/models"
)

var logger = slog.New(
	slug.NewHandler(slug.HandlerOptions{}, os.Stdout),
)

type publicErrorFormat struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func ErrorBuilder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := oops.
			Request(r, false)
		ctx := oops.WithBuilder(r.Context(), err)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func HandleHTTPError(ctx context.Context, w http.ResponseWriter, err error) {
	errOops, isOops := oops.AsOops(err)
	if !isOops {
		errOops, _ = oops.AsOops(
			oops.FromContext(ctx).Wrap(err),
		)
	}
	logger.Error(
		errOops.Error(),
		slog.Any("error", errOops),
	)

	w.Header().Set("Content-Type", "application/json")
	if statCode, found := errOops.Context()[models.HTTPCodeCtx]; found {
		w.WriteHeader(statCode.(int))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	pubErr := errOops.Public()
	if pubErr == "" {
		pubErr = "Internal Server Error"
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(publicErrorFormat{
		Error:   true,
		Message: pubErr,
	})
}

func SimpleHTTPError(
	ctx context.Context, w http.ResponseWriter, err error,
	msg string, httpCode int,
) {
	errB := oops.
		FromContext(ctx).
		Public(msg)
	if httpCode != 0 {
		errB = errB.With(models.HTTPCodeCtx, httpCode)
	}
	HandleHTTPError(ctx, w, errB.Wrap(err))
}
