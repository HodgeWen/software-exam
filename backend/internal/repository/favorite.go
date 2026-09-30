package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"software-exam/backend/internal/model"
)

// FavoriteWithQuestion 收藏列表行：收藏记录 + 对应题目（按用户隔离）
type FavoriteWithQuestion struct {
	Favorite model.Favorite
	Question model.Question
}

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Add 收藏：冲突时什么都不做，保证重复收藏不产生重复行（幂等）
func (r *FavoriteRepository) Add(userID, questionID uint) error {
	f := model.Favorite{UserID: userID, QuestionID: questionID}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&f).Error
}

// Delete 取消收藏；目标不存在时不报错（幂等）
func (r *FavoriteRepository) Delete(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.Favorite{}).Error
}

// List 收藏分页列表，按收藏时间倒序；Count 与取页分别执行，避免复用同一语句的子句残留
func (r *FavoriteRepository) List(userID uint, page, pageSize int) ([]FavoriteWithQuestion, int64, error) {
	filtered := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var fs []model.Favorite
	if err := filtered.Order("created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&fs).Error; err != nil {
		return nil, 0, err
	}
	if len(fs) == 0 {
		return nil, total, nil
	}

	// 第二跳批量取题目，按收藏顺序拼回列表行
	ids := make([]uint, len(fs))
	for i, f := range fs {
		ids[i] = f.QuestionID
	}
	var qs []model.Question
	if err := r.db.Find(&qs, ids).Error; err != nil {
		return nil, 0, err
	}
	byID := make(map[uint]model.Question, len(qs))
	for _, q := range qs {
		byID[q.ID] = q
	}
	rows := make([]FavoriteWithQuestion, 0, len(fs))
	for _, f := range fs {
		if q, ok := byID[f.QuestionID]; ok {
			rows = append(rows, FavoriteWithQuestion{Favorite: f, Question: q})
		}
	}
	return rows, total, nil
}
