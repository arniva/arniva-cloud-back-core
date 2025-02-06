package models

import "github.com/guregu/null/v5"

type BaseModel struct {
	ID        null.String `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt null.Time   `json:"zaman" gorm:"type:timestamp without time zone;default:now()"`
	UpdatedAt null.Time   `json:"guncelleme" gorm:"type:timestamp without time zone;default:now()"`
}
