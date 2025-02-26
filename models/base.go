package models

import (
	"time"
)

type BaseModel struct {
	ID        *string   `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp without time zone;default:now()"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp without time zone;default:now()"`
}
