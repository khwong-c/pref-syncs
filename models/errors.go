package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/samber/oops"
)

// Assigning context to HTTPCodeCtx presents HTTP Error Code in the response.
const HTTPCodeCtx = "HTTPCode"

const (
	ErrCodeNotFound = "NotFound"
)

func NewBadReqErr(
	ctx context.Context,
	err error,
	msg string,
	args ...any,
) error {
	if len(args) != 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if err == nil {
		err = errors.New(msg)
	}
	return oops.FromContext(ctx).
		With(HTTPCodeCtx, http.StatusBadRequest).
		Code(ErrCodeNotFound).
		Public(msg).
		Wrap(err)
}

func NewNotFoundErr(
	ctx context.Context,
	err error,
	msg string,
	args ...any,
) error {
	if len(args) != 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	if err == nil {
		err = errors.New(msg)
	}
	return oops.FromContext(ctx).
		With(HTTPCodeCtx, http.StatusNotFound).
		Code(ErrCodeNotFound).
		Public(msg).
		Wrap(err)
}

func IsNotFoundErr(err error) bool {
	e, ok := oops.AsOops(err)
	if !ok {
		return false
	}
	return e.Code() == ErrCodeNotFound
}
