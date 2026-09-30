package model

import "time"

// Chapter 章节属于科目；(SubjectID, Code) 构成稳定业务键，种子按它判重
type Chapter struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SubjectID uint      `gorm:"uniqueIndex:idx_chapters_subject_code;not null" json:"subject_id"`
	Code      string    `gorm:"uniqueIndex:idx_chapters_subject_code;size:64;not null" json:"code"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Sort      int       `gorm:"not null" json:"sort"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
