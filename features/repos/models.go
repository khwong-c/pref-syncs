package repos

import (
	"time"
	"uuid"
)

type App struct {
	ID   uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	Name string
	Desc string
}

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	AuthProvider string    `gorm:"index:user_auth_idx,priority:1"`
	AuthUserID   string    `gorm:"index:user_auth_idx,priority:2"`
	Apps         []*App    `gorm:"many2many:user_apps;"`
}

type PrefEntry struct {
	AppID     uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	UserID    uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	UpdatedAt time.Time
	Payload   string
}
