package service

import (
	"fmt"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

// 列表分页默认值与上限：防止无界查询
const (
	mistakePageDefault = 20
	mistakePageMax     = 100
)

// MistakeService 错题本：列表（按科目过滤 + 分页）、重刷取题、手动移除；全部按用户隔离
type MistakeService struct {
	repo *repository.MistakeRepository
}

func NewMistakeService(repo *repository.MistakeRepository) *MistakeService {
	return &MistakeService{repo: repo}
}

func (s *MistakeService) List(userID, subjectID uint, page, pageSize int) ([]repository.MistakeWithQuestion, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = mistakePageDefault
	}
	if pageSize > mistakePageMax {
		pageSize = mistakePageMax
	}
	return s.repo.List(userID, subjectID, page, pageSize)
}

// Questions 错题重刷题源：返回当前用户错题题目序列（不含正确答案，判分仍走提交接口）
func (s *MistakeService) Questions(userID uint) ([]model.Question, error) {
	qs, err := s.repo.ListQuestions(userID)
	if err != nil {
		return nil, fmt.Errorf("查询错题题源: %w", err)
	}
	return qs, nil
}

// Remove 手动移除错题；不存在的错题视为已移除（幂等）
func (s *MistakeService) Remove(userID, questionID uint) error {
	if err := s.repo.Delete(userID, questionID); err != nil {
		return fmt.Errorf("移除错题: %w", err)
	}
	return nil
}
