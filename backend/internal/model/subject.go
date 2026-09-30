package model

import "time"

// Subject 科目；Code 为稳定业务编码（如 soft-designer），种子导入与跨表引用一律走 Code，不写死单一科目
type Subject struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
