package features

import (
	"github.com/samber/do/v2"
	"github.com/samber/oops"

	"github.com/khwong-c/pref-syncs/features/repos"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

type AppLogic struct {
	dataRepo         *repos.DataRepo
	notificationRepo repos.NotificationRepoI
}

func NewAppLogic(inj do.Injector) *AppLogic {
	app := &AppLogic{
		dataRepo:         repos.NewDataRepo(inj),
		notificationRepo: di.InvokeOrProvide(inj, repos.NewSingleContainerNotifier),
	}
	if err := app.dataRepo.InitDB(); err != nil {
		panic(oops.Wrapf(err, "failed to initialize database"))
	}
	return app
}
