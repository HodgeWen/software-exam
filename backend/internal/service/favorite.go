package service

import (
	"fmt"

	"software-exam/backend/internal/repository"
)

// 列表分页默认值与上限：防止无界查询
const (
	favoritePageDefault = 20
	favoritePageMax     = 100
)

// FavoriteService 收藏：幂等收藏/取消、分页列表；全部按用户隔离
type FavoriteService struct {
	repo *repository.FavoriteRepository
}

func NewFavoriteService(repo *repository.FavoriteRepository) *FavoriteService {
	return &FavoriteService{repo: repo}
}

// Add 收藏；重复收藏不产生重复记录（幂等）
func (s *FavoriteService) Add(userID, questionID uint) error {
	if err := s.repo.Add(userID, questionID); err != nil {
		return fmt.Errorf("收藏题目: %w", err)
	}
	return nil
}

// Remove 取消收藏；未收藏也视为成功（幂等）
func (s *FavoriteService) Remove(userID, questionID uint) error {
	if err := s.repo.Delete(userID, questionID); err != nil {
		return fmt.Errorf("取消收藏: %w", err)
	}
	return nil
}

// List 收藏分页列表（响应含 total），按用户隔离
func (s *FavoriteService) List(userID uint, page, pageSize int) ([]repository.FavoriteWithQuestion, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = favoritePageDefault
	}
	if pageSize > favoritePageMax {
		pageSize = favoritePageMax
	}
	return s.repo.List(userID, page, pageSize)
}
