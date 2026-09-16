package features

import (
	"context"
	"uuid"

	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/features/repos"
)

func (a *AppLogic) UpdatePref(ctx context.Context, pref *repos.PrefEntry, src *string) (*repos.PrefEntry, error) {
	newPref, err := a.dataRepo.PutPreference(ctx, pref.UserID, pref.AppID, pref.Payload)
	if err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	if err := a.notificationRepo.Notify(ctx, &repos.Notification{
		User: pref.UserID,
		App:  pref.AppID,
		Src:  src,
	}); err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return newPref, nil
}

func (a *AppLogic) GetPref(ctx context.Context, user uuid.UUID, app uuid.UUID) (*repos.PrefEntry, error) {
	return a.dataRepo.GetPreference(ctx, user, app)
}

func (a *AppLogic) DeletePref(ctx context.Context, user uuid.UUID, app uuid.UUID) error {
	return a.dataRepo.DeletePreference(ctx, user, app)
}
