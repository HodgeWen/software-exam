package model

import "time"

// 题型：单选答案恰 1 个标号，多选答案为标号集合（判对要求所选集合与正确集合完全一致）
const (
	QuestionTypeSingle   = "single"
	QuestionTypeMultiple = "multiple"
)

// Option 选择题选项：Key 为选项标号（A/B/C/D…），Text 为选项内容
type Option struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Question 题目；Code 为稳定业务键（科目编码 + 章节或试卷编码 + 题号），种子按它判重；
// ChapterID 可空——真题试卷题不属于任何章节；Answer 只用于判分，默认不随模型序列化外泄
type Question struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SubjectID uint      `gorm:"index;not null" json:"subject_id"`
	ChapterID *uint     `gorm:"index" json:"chapter_id"`
	Code      string    `gorm:"uniqueIndex;size:128;not null" json:"code"`
	No        int       `gorm:"not null" json:"no"`
	Type      string    `gorm:"size:16;not null" json:"type"`
	Stem      string    `gorm:"type:text;not null" json:"stem"`
	Options   []Option  `gorm:"serializer:json" json:"options"`
	Answer    []string  `gorm:"serializer:json" json:"-"`
	Analysis  string    `gorm:"type:text;not null" json:"analysis"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
