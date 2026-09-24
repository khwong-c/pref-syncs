package repos

import (
	"time"
	"uuid"
)

type App struct {
	ID            uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	Name          string
	Desc          string
	CreatedBy     uuid.UUID
	CreatedByUser *User `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MaxPayload    int
	MaxUser       int
}

type AppUpdate struct {
	Name       *string
	Desc       *string
	MaxPayload *int
	MaxUser    *int
}

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	AuthProvider string    `gorm:"index:user_auth_idx,priority:1"`
	AuthUserID   string    `gorm:"index:user_auth_idx,priority:2"`
	Apps         []*App    `gorm:"many2many:user_apps;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsAdmin      bool
}

type PrefEntry struct {
	AppID     uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	UserID    uuid.UUID `gorm:"primaryKey;autoIncrement:false"`
	App       *App      `gorm:"foreignKey:AppID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User      *User     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UpdatedAt time.Time
	Payload   string
}
