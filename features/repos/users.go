package repos

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/samber/oops"
	"gorm.io/gorm"

	"github.com/khwong-c/pref-syncs/drivers/sql"
	"github.com/khwong-c/pref-syncs/models"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type repoCtxKey string

var (
	gormDBCtxKey repoCtxKey = "gorm.db"
)

type DataRepo struct {
	db *gorm.DB
}

func NewDataRepo(injector do.Injector) *DataRepo {
	return &DataRepo{
		db: di.InvokeOrProvide(injector, sql.NewInMemorySQLite),
	}
}

func (r *DataRepo) getTxFromCtx(ctx context.Context) *gorm.DB {
	if tx, found := ctx.Value(&gormDBCtxKey).(*gorm.DB); found {
		return tx
	}
	return r.db
}

func (r *DataRepo) withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, &gormDBCtxKey, tx)
}

func (r *DataRepo) InitDB() error {
	tx := r.db
	if err := tx.AutoMigrate(
		&App{},
		&User{},
		&PrefEntry{},
	); err != nil {
		return oops.Wrap(err)
	}
	return nil
}

func (r *DataRepo) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	tx := r.getTxFromCtx(ctx)
	user, err := gorm.G[*User](tx).
		Where(&User{ID: id}).
		Preload("Apps", nil).
		Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.NewNotFoundErr(
				ctx, err,
				"user not found: %s", id,
			)
		}
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return user, nil
}

func (r *DataRepo) GetOrCreateUser(ctx context.Context, issuer string, subject string) (*User, error) {
	tx := r.getTxFromCtx(ctx)
	var user *User
	err := tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)
		userRecord, err := gorm.G[User](tx).
			Where(&User{
				AuthProvider: issuer,
				AuthUserID:   subject,
			}).
			Take(ctx)
		if err == nil {
			user = &userRecord
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return oops.FromContext(ctx).Wrap(err)
		}

		user = &User{
			ID:           uuid.NewV7(),
			AuthProvider: issuer,
			AuthUserID:   subject,
		}
		err = gorm.G[User](tx).Create(ctx, user)
		if err != nil {
			return oops.FromContext(ctx).Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, err
}

func (r *DataRepo) CreateUser(ctx context.Context, user *User) (*User, error) {
	tx := r.getTxFromCtx(ctx)
	if err := gorm.G[User](tx).Create(ctx, user); err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return user, nil
}

func (r *DataRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	tx := r.getTxFromCtx(ctx)
	return tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)
		user, err := r.GetUser(ctx, id)
		if err != nil {
			return err
		}

		// Remove user.
		if err := tx.Delete(&user).Error; err != nil {
			return oops.FromContext(ctx).Wrap(err)
		}
		return nil
	})
}

func (r *DataRepo) AuthorizeUserToApp(ctx context.Context, uid uuid.UUID, aid uuid.UUID) error {
	tx := r.getTxFromCtx(ctx)
	err := tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)
		user, err := r.GetUser(ctx, uid)
		if err != nil {
			return err
		}

		if _, err := r.GetApp(ctx, aid); err != nil {
			return err
		}
		if lo.ContainsBy(
			user.Apps,
			func(app *App) bool { return app.ID == aid },
		) {
			return nil
		}

		user.Apps = append(
			lo.Map(
				user.Apps,
				func(app *App, _ int) *App { return &App{ID: app.ID} },
			),
			&App{ID: aid},
		)

		if _, err := gorm.G[*User](tx).Where(&User{ID: uid}).Updates(ctx, user); err != nil {
			return oops.
				FromContext(ctx).
				Public(fmt.Sprintf("Failed to authorize user to app: %s", aid)).
				Wrap(err)
		}
		return nil
	})
	return err
}

func (r *DataRepo) DeauthorizeUserFromApp(ctx context.Context, uid uuid.UUID, aid uuid.UUID) error {
	tx := r.getTxFromCtx(ctx)
	err := tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)
		user, err := r.GetUser(ctx, uid)
		if err != nil {
			return err
		}
		if _, err := r.GetApp(ctx, aid); err != nil {
			return err
		}

		if !lo.ContainsBy(
			user.Apps, func(app *App) bool {
				return app.ID == aid
			},
		) {
			return nil
		}

		user.Apps = lo.RejectMap(
			user.Apps,
			func(app *App, index int) (*App, bool) {
				return &App{ID: app.ID}, app.ID == aid
			},
		)
		if _, err := gorm.G[*User](tx).Where(&User{ID: uid}).Updates(ctx, user); err != nil {
			return oops.
				FromContext(ctx).
				Public(fmt.Sprintf("Unable to deauthorized to app: %s", aid)).
				Wrap(err)
		}
		return nil
	})
	return err
}

func (r *DataRepo) IsUserAuthorizedToApp(ctx context.Context, uid uuid.UUID, aid uuid.UUID) (bool, error) {
	tx := r.getTxFromCtx(ctx)
	count := int64(0)
	err := tx.
		WithContext(ctx).
		Table("user_apps").
		Where(
			"user_id = ? AND app_id = ?",
			uid, aid,
		).
		Count(&count).Error
	if err != nil {
		return false, oops.FromContext(ctx).Wrap(err)
	}
	return count > 0, nil
}

func (r *DataRepo) PromoteUserToAdmin(ctx context.Context, uid uuid.UUID) (*User, error) {
	tx := r.getTxFromCtx(ctx)
	var user *User
	err := tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)

		var err error
		user, err = r.GetUser(ctx, uid)
		if err != nil {
			return err
		}

		user.IsAdmin = true
		if _, err := gorm.G[*User](tx).Updates(ctx, user); err != nil {
			return oops.
				FromContext(ctx).
				Public(fmt.Sprintf("Unable to promote user to admin: %s", uid)).
				Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}
