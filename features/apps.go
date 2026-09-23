package features

import (
	"context"
	"net/http"
	"uuid"

	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/models"
)

const (
	defaultMaxPayload = 4096 // 4-KB of Payload
	defaultMaxUser    = 10   // 10 Users
)

type AppUpsertReq struct {
	ID         uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	Name       *string
	Desc       *string
	CreatedBy  uuid.UUID
	MaxPayload *int
	MaxUser    *int
}

func (a *AppLogic) CreateApp(ctx context.Context, req *AppUpsertReq) (*repos.App, error) {
	// Validate Fields
	if req.Name == nil {
		return nil, models.NewBadReqErr(ctx, nil, "name is required")
	}
	if req.Desc == nil {
		return nil, models.NewBadReqErr(ctx, nil, "desc is required")
	}

	// Fill the default values
	if req.MaxPayload == nil {
		req.MaxPayload = new(defaultMaxPayload)
	}
	if req.MaxUser == nil {
		req.MaxUser = new(defaultMaxUser)
	}

	// Validate App Options
	if *req.MaxPayload != defaultMaxPayload {
		return nil, models.NewBadReqErr(ctx, nil, "max payload must be %d", defaultMaxPayload)
	}
	if *req.MaxUser != defaultMaxUser {
		return nil, models.NewBadReqErr(ctx, nil, "max user must be %d", defaultMaxUser)
	}

	newEntry := &repos.App{
		ID:         uuid.NewV7(),
		CreatedBy:  req.CreatedBy,
		Name:       *req.Name,
		Desc:       *req.Desc,
		MaxPayload: *req.MaxPayload,
		MaxUser:    *req.MaxUser,
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

func (a *AppLogic) ModifyApp(
	ctx context.Context,
	userID uuid.UUID,
	appID uuid.UUID,
	req *AppUpsertReq,
) (*repos.App, error) {
	// Validate Request
	if req.Name != nil && *req.Name == "" {
		return nil, models.NewBadReqErr(ctx, nil, "name cannot be empty")
	}
	if req.Desc != nil && *req.Desc == "" {
		return nil, models.NewBadReqErr(ctx, nil, "desc cannot be empty")
	}
	if req.MaxPayload != nil && *req.MaxPayload != defaultMaxPayload {
		return nil, models.NewBadReqErr(ctx, nil, "max payload must be %d", defaultMaxPayload)
	}
	if req.MaxUser != nil && *req.MaxUser != defaultMaxUser {
		return nil, models.NewBadReqErr(ctx, nil, "max user must be %d", defaultMaxUser)
	}

	// Check Ownership
	app, err := a.dataRepo.GetApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app.CreatedBy != userID {
		return nil, oops.FromContext(ctx).
			Public("Unauthorized").
			With(models.HTTPCodeCtx, http.StatusForbidden).
			Errorf(
				"User %s attempted to modify app %s, owned by %s",
				userID,
				appID,
				app.CreatedBy,
			)
	}

	app, err = a.dataRepo.ModifyApp(
		ctx, appID,
		&repos.AppUpdate{
			Name:       req.Name,
			Desc:       req.Desc,
			MaxPayload: req.MaxPayload,
			MaxUser:    req.MaxUser,
		},
	)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (a *AppLogic) DeleteApp(ctx context.Context, id uuid.UUID) error {
	return a.dataRepo.DeleteApp(ctx, id)
}
