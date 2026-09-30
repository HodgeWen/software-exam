package repository

import (
	"software-exam/backend/internal/model"

	"gorm.io/gorm"
)

type ExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) *ExamRepository {
	return &ExamRepository{db: db}
}

func (r *ExamRepository) Create(exam *model.Exam) error {
	return r.db.Create(exam).Error
}

// FindByUser 按 (id, user_id) 取考试记录：他人记录与不存在一律 ErrRecordNotFound，
// 业务数据按用户隔离不暴露存在性
func (r *ExamRepository) FindByUser(userID, examID uint) (*model.Exam, error) {
	var exam model.Exam
	if err := r.db.Where("user_id = ? AND id = ?", userID, examID).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *ExamRepository) Save(exam *model.Exam) error {
	return r.db.Save(exam).Error
}
