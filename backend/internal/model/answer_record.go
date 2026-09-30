package model

import "time"

// AnswerRecord 每次提交落一条的答题记录；业务数据一律按用户隔离，统计基于它聚合
type AnswerRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	QuestionID uint      `gorm:"index;not null" json:"question_id"`
	Selected   []string  `gorm:"serializer:json;not null" json:"selected"`
	Correct    bool      `gorm:"not null" json:"correct"`
	CreatedAt  time.Time `json:"created_at"`
}
