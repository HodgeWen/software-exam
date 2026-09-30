package repository

import (
	"software-exam/backend/internal/model"

	"gorm.io/gorm"
)

type AnswerRecordRepository struct {
	db *gorm.DB
}

func NewAnswerRecordRepository(db *gorm.DB) *AnswerRecordRepository {
	return &AnswerRecordRepository{db: db}
}

func (r *AnswerRecordRepository) Create(rec *model.AnswerRecord) error {
	return r.db.Create(rec).Error
}

// AnswerDim 统计聚合维度行：一次作答经题目归属到的科目/章节与对错；
// 真题试卷题不属任何章节，ChapterID/ChapterName 为空
type AnswerDim struct {
	SubjectID   uint
	SubjectName string
	ChapterID   *uint
	ChapterName *string
	Correct     bool
}

// ListDims 取当前用户全部作答的维度行，供 service 层聚合成统计；按用户隔离
func (r *AnswerRecordRepository) ListDims(userID uint) ([]AnswerDim, error) {
	var dims []AnswerDim
	err := r.db.Model(&model.AnswerRecord{}).
		Select("questions.subject_id", "subjects.name AS subject_name",
			"questions.chapter_id", "chapters.name AS chapter_name", "answer_records.correct").
		Joins("JOIN questions ON questions.id = answer_records.question_id").
		Joins("LEFT JOIN subjects ON subjects.id = questions.subject_id").
		Joins("LEFT JOIN chapters ON chapters.id = questions.chapter_id").
		Where("answer_records.user_id = ?", userID).
		Scan(&dims).Error
	return dims, err
}
