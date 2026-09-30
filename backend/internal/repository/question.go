package repository

import (
	"software-exam/backend/internal/model"

	"gorm.io/gorm"
)

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// FindByID 取单个题目（含正确答案与解析，仅供服务端判分使用）
func (r *QuestionRepository) FindByID(id uint) (*model.Question, error) {
	var q model.Question
	if err := r.db.First(&q, id).Error; err != nil {
		return nil, err
	}
	return &q, nil
}
