package model

import "time"

// Paper 真题试卷属于科目；Code 为稳定业务编码（科目内唯一），种子按它判重
type Paper struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SubjectID       uint      `gorm:"index;not null" json:"subject_id"`
	Code            string    `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name            string    `gorm:"size:128;not null" json:"name"`
	DurationMinutes int       `gorm:"not null" json:"duration_minutes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PaperQuestion 试卷-题目关联；复合主键天然防重复，No 为卷内题号
type PaperQuestion struct {
	PaperID    uint `gorm:"primaryKey" json:"paper_id"`
	QuestionID uint `gorm:"primaryKey" json:"question_id"`
	No         int  `gorm:"not null" json:"no"`
}
