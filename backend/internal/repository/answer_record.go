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
