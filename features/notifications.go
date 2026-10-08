package features

import (
	"context"
	"uuid"

	"github.com/samber/oops"
	"github.com/samber/ro"

	"github.com/khwong-c/pref-syncs/features/repos"
)

func (a *AppLogic) SubscribeNotification(ctx context.Context, user, app uuid.UUID, src *string) (ro.Observable[*repos.Notification], ro.Teardown, error) {
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
