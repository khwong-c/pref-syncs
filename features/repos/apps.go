package repos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"uuid"

	"github.com/samber/oops"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/models"
)

func (r *DataRepo) GetApp(ctx context.Context, id uuid.UUID) (*App, error) {
	app, err := gorm.G[*App](r.db).
		Where(&App{ID: id}).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			e := oops.
				FromContext(ctx).
				Public(fmt.Sprintf("app not found: %s", id)).
				With(models.HTTPCodeCtx, http.StatusNotFound).
				Wrap(err)
			return nil, e
		}
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return app, nil
}

func (r *DataRepo) CreateApp(ctx context.Context, app *App) (*App, error) {
	if err := gorm.G[App](r.db).Create(ctx, app); err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return app, nil
}

func (r *DataRepo) ModifyApp(ctx context.Context, id uuid.UUID, app *App) (*App, error) {
	rows, err := gorm.G[*App](r.db).Where(&App{ID: id}).Updates(ctx, app)
	if err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	if rows == 0 {
		return nil, oops.
			FromContext(ctx).
			Public(fmt.Sprintf("App not found: %s", id)).
			With(models.HTTPCodeCtx, http.StatusNotFound).
			Wrap(err)
	}
	return app, nil
}

func (r *DataRepo) DeleteApp(ctx context.Context, id uuid.UUID) error {
	rows, err := gorm.G[App](r.db).Where(&App{ID: id}).Delete(ctx)
	if err != nil {
		return oops.FromContext(ctx).Wrap(err)
	}
	if rows == 0 {
		return oops.
			FromContext(ctx).
			Public(fmt.Sprintf("App not found: %s", id)).
			With(models.HTTPCodeCtx, http.StatusNotFound).
			Wrap(err)
	}
	return nil
}
