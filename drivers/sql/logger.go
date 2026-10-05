package sql

import (
	"log"
	"os"

	"gorm.io/gorm/logger"
)

func newGormLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		Colorful:                  true,
		IgnoreRecordNotFoundError: true,
	})
}
