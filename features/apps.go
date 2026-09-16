package features

import (
	"context"
	"uuid"

	"github.com/khwong-c/pref-syncs/features/repos"
)

func (a *AppLogic) CreateApp(ctx context.Context, app *repos.App) (*repos.App, error) {
	newEntry := &repos.App{
		ID:   uuid.NewV7(),
		Name: app.Name,
		Desc: app.Desc,
	}
	newEntry, err := a.dataRepo.CreateApp(ctx, newEntry)
	if err != nil {
		return nil, err
	}
	return newEntry, nil
}

func (a *AppLogic) GetApp(ctx context.Context, id uuid.UUID) (*repos.App, error) {
	entry, err := a.dataRepo.GetApp(ctx, id)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (a *AppLogic) ModifyApp(ctx context.Context, id uuid.UUID, app *repos.App) (*repos.App, error) {
	app, err := a.dataRepo.ModifyApp(ctx, id, app)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (a *AppLogic) DeleteApp(ctx context.Context, id uuid.UUID) error {
	return a.dataRepo.DeleteApp(ctx, id)
}
