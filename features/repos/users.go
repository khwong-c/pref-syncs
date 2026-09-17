package repos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
	id1 := uuid.MustParse("01a0a690-d36b-7297-85d6-13ca9f415cf3")
	if err := gorm.G[User](tx).Create(context.Background(), &User{
		ID:           uuid.MustParse("01a0734e-795a-72bc-874b-9b90c41ddc8b"),
		AuthProvider: "",
		AuthUserID:   "",
		Apps: []*App{
			{
				ID:   id1,
				Name: "App 1",
				Desc: "App 1",
			},
			{
				ID:   uuid.NewV4(),
				Name: "App 1",
				Desc: "App 1",
			},
		},
	}); err != nil {
		return oops.Wrap(err)
	}
	if err := gorm.G[User](tx).Create(context.Background(), &User{
		ID:           uuid.MustParse("01a08408-6237-7008-9508-263f5cf27759"),
		AuthProvider: "",
		AuthUserID:   "",
		Apps: []*App{
			{ID: id1},
		},
	}); err != nil {
		return oops.Wrap(err)
	}
	//r.PutReference(context.Background(), uuid.MustParse("01a0734e-795a-72bc-874b-9b90c41ddc8b"), uuid.MustParse("01a08408-6237-7008-9508-263f5cf27759"), &User{})
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
			e := oops.
				FromContext(ctx).
				Public(fmt.Sprintf("user not found: %s", id)).
				With(models.HTTPCodeCtx, http.StatusNotFound).
				Wrap(err)
			return nil, e
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
	rows, err := gorm.G[User](tx).Where(&User{ID: id}).Delete(ctx)
	if err != nil {
		return oops.FromContext(ctx).Wrap(err)
	}
	if rows == 0 {
		return oops.
			FromContext(ctx).
			Public(fmt.Sprintf("user not found: %s", id)).
			With(models.HTTPCodeCtx, http.StatusNotFound).
			Wrap(err)
	}
	return nil
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
