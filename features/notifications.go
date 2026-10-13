package features

import (
	"context"
	"uuid"

	"github.com/samber/oops"
	"github.com/samber/ro"

	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/models"
)

func (a *AppLogic) SubscribeNotification(ctx context.Context, user, app uuid.UUID, src *string) (ro.Observable[*repos.Notification], ro.Teardown, error) {
	if _, err := a.dataRepo.GetApp(ctx, app); err != nil {
		return nil, nil, oops.FromContext(ctx).Wrap(err)
	}
	authorized, err := a.dataRepo.IsUserAuthorizedToApp(ctx, user, app)
	if err != nil {
		return nil, nil, oops.FromContext(ctx).Wrap(err)
	}
	if !authorized {
		return nil, nil, models.NewUnauthorizedErr(
			ctx, nil, "user is not authorized to app",
		)
	}
	obs, teardown, err := a.notificationRepo.Subscribe(ctx, user, app)
	if err != nil {
		return nil, nil, oops.FromContext(ctx).Wrap(err)
	}
	return ro.Pipe1(
		obs,
		ro.Filter(func(item *repos.Notification) bool {
			return src == nil || item.Src == nil || *item.Src != *src
		}),
	), teardown, nil
}
