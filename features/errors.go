package features

import (
	"context"
	"fmt"
	"net/http"

	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/models"
)

func newBadReqErr(
	ctx context.Context,
	msg string,
	args ...any,
) error {
	if len(args) == 0 {
		return oops.FromContext(ctx).
			With(models.HTTPCodeCtx, http.StatusBadRequest).
			Public(msg).
			New(msg)
	}
	errMsg := fmt.Sprintf(msg, args...)
	return oops.FromContext(ctx).
		With(models.HTTPCodeCtx, http.StatusBadRequest).
		Public(errMsg).
		New(errMsg)
}
