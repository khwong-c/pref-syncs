package features

import (
	"context"
	"uuid"

	"github.com/samber/oops"
	"github.com/samber/ro"

	"github.com/khwong-c/pref-syncs/features/repos"
)

func (a *AppLogic) GetNotificationStream(ctx context.Context, user, app uuid.UUID, src *string) (ro.Observable[*repos.Notification], error) {
	stream, err := a.notificationRepo.GetStream(ctx, user, app)
	if err != nil {
		return nil, oops.FromContext(ctx).Wrap(err)
	}
	return ro.Pipe1(
		stream,
		ro.Filter[*repos.Notification](
			func(item *repos.Notification) bool {
				return src == nil || *item.Src != *src
			},
		),
	), nil
}
