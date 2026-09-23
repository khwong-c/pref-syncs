package repos

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/samber/oops"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/khwong-c/pref-syncs/models"
)

func (r *DataRepo) PutPreference(ctx context.Context, uid uuid.UUID, aid uuid.UUID, payload string) (*PrefEntry, error) {
	tx := r.getTxFromCtx(ctx)
	prefEntry := &PrefEntry{
		AppID:     aid,
		UserID:    uid,
		UpdatedAt: time.Now(),
		Payload:   payload,
	}
	if err := gorm.G[PrefEntry](
		tx, clause.OnConflict{UpdateAll: true},
	).Create(ctx, prefEntry); err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return prefEntry, nil
}

func (r *DataRepo) GetPreference(ctx context.Context, uid uuid.UUID, aid uuid.UUID) (*PrefEntry, error) {
	tx := r.getTxFromCtx(ctx)
	prefEntry := &PrefEntry{
		AppID:  aid,
		UserID: uid,
	}
	entryFound, err := gorm.G[PrefEntry](tx).Where(prefEntry).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.NewNotFoundErr(
				ctx, err,
				"preference not found: %s, %s", uid, aid,
			)
		}
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return &entryFound, nil
}

func (r *DataRepo) DeletePreference(ctx context.Context, uid uuid.UUID, aid uuid.UUID) error {
	tx := r.getTxFromCtx(ctx)
	prefEntry := &PrefEntry{
		AppID:  aid,
		UserID: uid,
	}
	err := tx.Transaction(func(tx *gorm.DB) error {
		ctx = r.withTx(ctx, tx)
		entryFound, err := gorm.G[PrefEntry](tx).Where(prefEntry).Count(ctx, "*")
		if err != nil {
			return oops.FromContext(ctx).Wrap(err)
		}
		if entryFound == 0 {
			return models.NewNotFoundErr(ctx, err, "preference not found: %s, %s", uid, aid)
		}
		_, err = gorm.G[PrefEntry](tx).Where(prefEntry).Delete(ctx)
		return oops.FromContext(ctx).Wrap(err)
	})
	return err
}
