package model

import "time"

// Favorite 收藏条目：(UserID, QuestionID) 唯一——重复收藏不产生重复行
type Favorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"uniqueIndex:idx_favorites_user_question;not null" json:"user_id"`
	QuestionID uint      `gorm:"uniqueIndex:idx_favorites_user_question;not null" json:"question_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
