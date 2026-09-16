package features

import (
	"context"
	"uuid"

	"github.com/khwong-c/pref-syncs/features/repos"
)

func (a *AppLogic) GetUser(ctx context.Context, id uuid.UUID) (*repos.User, error) {
	user, err := a.dataRepo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (a *AppLogic) CreateUser(ctx context.Context, user *repos.User) (*repos.User, error) {
	newUser := &repos.User{
		ID:           uuid.NewV7(),
		AuthProvider: user.AuthProvider,
		AuthUserID:   user.AuthUserID,
	}
	newUser, err := a.dataRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (a *AppLogic) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := a.dataRepo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (a *AppLogic) AuthoriseUserToApp(ctx context.Context, uid, appID uuid.UUID) (*repos.User, error) {
	if err := a.dataRepo.AuthorizeUserToApp(ctx, uid, appID); err != nil {
		return nil, err
	}
	return a.dataRepo.GetUser(ctx, uid)
}

func (a *AppLogic) DeauthoriseUserToApp(ctx context.Context, uid uuid.UUID, appID uuid.UUID) (*repos.User, error) {
	if err := a.dataRepo.DeauthorizeUserFromApp(ctx, uid, appID); err != nil {
		return nil, err
	}
	return a.dataRepo.GetUser(ctx, uid)
}
