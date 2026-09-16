package main

import (
	"syscall"

	"github.com/samber/do/v2"

	"github.com/khwong-c/pref-syncs/server"
	"github.com/khwong-c/pref-syncs/tooling/di"
)

func main() {
	i := do.New()
	//_ = di.InvokeOrProvide(i, func(injector do.Injector) (*gorm.DB, error) {
	//	return sql.NewSQLite("tmp.db")
	//})
	s := di.InvokeOrProvide(i, server.NewServer)

	go s.ListenAndServe() //nolint:errcheck
	//go func() {
	//	time.Sleep(3 * time.Second)
	//	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	//}()
	// Shutdown Gracefully
	_, _ = i.ShutdownOnSignals(
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL,
	)
}
