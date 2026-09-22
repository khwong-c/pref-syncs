package features

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/features/repos"
)

func (*AppLogic) ParsePayloadWithLimit(ctx context.Context, r io.Reader, app *repos.App) (json.RawMessage, error) {
	lr := io.LimitReader(r, int64(app.MaxPayload))
	raw := json.RawMessage{}
	err := json.NewDecoder(lr).Decode(&raw)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, newBadReqErr(ctx, "payload exceeds maximum size")
	}
	if err != nil {
		return nil, newBadReqErr(ctx, err.Error())
	}
	return raw, nil
}

func (*AppLogic) CheckUserLimit(ctx context.Context, app *repos.App) error {
	return oops.FromContext(ctx).Wrap(errors.ErrUnsupported)
}
