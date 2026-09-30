package model

import "time"

// 考试状态：开始考试生成进行中记录，交卷后一次性流转为已提交，不可重复交卷
const (
	ExamStatusInProgress = "in_progress"
	ExamStatusSubmitted  = "submitted"
)

// Exam 模拟考试记录：整卷开始/交卷时间与得分快照（答对数/总题数/正确率）；
// 逐题作答不在此表冗余存储，统一落 AnswerRecord（答错经同一入口入错题本）
type Exam struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index;not null" json:"user_id"`
	PaperID      uint       `gorm:"index;not null" json:"paper_id"`
	Status       string     `gorm:"size:16;not null" json:"status"`
	StartedAt    time.Time  `gorm:"not null" json:"started_at"`
	SubmittedAt  *time.Time `json:"submitted_at"`
	CorrectCount int        `gorm:"not null" json:"correct_count"`
	TotalCount   int        `gorm:"not null" json:"total_count"`
	Accuracy     float64    `gorm:"not null" json:"accuracy"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
