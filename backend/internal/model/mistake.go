package model

import "time"

// Mistake 错题本条目：(UserID, QuestionID) 唯一——重复答错不产生重复行，只累加错误次数
type Mistake struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex:idx_mistakes_user_question;not null" json:"user_id"`
	QuestionID  uint      `gorm:"uniqueIndex:idx_mistakes_user_question;not null" json:"question_id"`
	WrongCount  int       `gorm:"not null" json:"wrong_count"`
	LastWrongAt time.Time `gorm:"not null" json:"last_wrong_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
