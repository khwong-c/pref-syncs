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

func (r *DataRepo) ModifyApp(
	ctx context.Context, id uuid.UUID, newEntry *AppUpdate,
) (*App, error) {
	app := &App{
		Name:       "",
		Desc:       "",
		MaxPayload: 0,
		MaxUser:    0,
	}
	if newEntry.Name != nil {
		app.Name = *newEntry.Name
	}
	if newEntry.Desc != nil {
		app.Desc = *newEntry.Desc
	}
	if newEntry.MaxPayload != nil {
		app.MaxPayload = *newEntry.MaxPayload
	}
	if newEntry.MaxUser != nil {
		app.MaxUser = *newEntry.MaxUser
	}

	tx := r.getTxFromCtx(ctx)
	rows, err := gorm.G[*App](tx).Where(&App{ID: id}).Updates(ctx, app)
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
	tx := r.getTxFromCtx(ctx)
	err := tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)

		// Remove All Preferences from the App
		appIDCond := &PrefEntry{AppID: id}
		if _, err := gorm.G[PrefEntry](tx).Where(appIDCond).Delete(ctx); err != nil {
			return oops.FromContext(ctx).Wrap(err)
		}

		// Disassociate all Users from the App
		if err := tx.
			WithContext(ctx).
			Table("user_apps").
			Where(
				"app_id = ?",
				id,
			).Delete(nil).Error; err != nil {
			return oops.FromContext(ctx).Wrap(err)
		}

		rows, err := gorm.G[App](tx).Where(&App{ID: id}).Delete(ctx)
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
	})
	return err

}
