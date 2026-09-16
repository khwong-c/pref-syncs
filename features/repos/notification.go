package repos

import (
	"context"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/samber/ro"
)

type Notification struct {
	User uuid.UUID
	App  uuid.UUID
	Src  *string
}

type NotificationRepoI interface {
	Notify(ctx context.Context, notification *Notification) error
	GetStream(ctx context.Context, uid uuid.UUID, aid uuid.UUID) (ro.Observable[*Notification], error)
}

type SingleContainerNotifier struct {
	stream ro.Subject[*Notification]
}

func NewSingleContainerNotifier(injector do.Injector) (*SingleContainerNotifier, error) {
	return &SingleContainerNotifier{
		stream: ro.NewPublishSubject[*Notification](),
	}, nil
}

func (r *SingleContainerNotifier) Notify(ctx context.Context, notification *Notification) error {
	r.stream.NextWithContext(ctx, notification)
	return nil
}

// GetStream returns a filtered stream of notifications for a given user, app. Event from the specified source is excluded.
func (r *SingleContainerNotifier) GetStream(ctx context.Context, uid uuid.UUID, aid uuid.UUID) (ro.Observable[*Notification], error) {
	return ro.Pipe1(
		r.stream,
		ro.Filter(func(notification *Notification) bool {
			return notification.User == uid && notification.App == aid
		}),
	), nil
}
