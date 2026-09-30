package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

func newFavoriteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("打开内存 SQLite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("取底层 sql.DB: %v", err)
	}
	// 内存库每个连接独立，限制单连接避免「no such table」
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.Favorite{}); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	return db
}

// TestFavoriteAddIdempotent 收藏幂等：重复收藏不产生重复记录，不同用户/不同题目各自独立
func TestFavoriteAddIdempotent(t *testing.T) {
	db := newFavoriteTestDB(t)
	repo := NewFavoriteRepository(db)

	for _, call := range []struct{ user, question uint }{
		{1, 100},
		{1, 100}, // 同用户同题重复收藏
		{1, 100}, // 再重复
		{2, 100}, // 其他用户
		{1, 200}, // 同用户不同题
	} {
		if err := repo.Add(call.user, call.question); err != nil {
			t.Fatalf("Add(%d, %d): %v", call.user, call.question, err)
		}
	}

	var fs []model.Favorite
	if err := db.Order("user_id, question_id").Find(&fs).Error; err != nil {
		t.Fatalf("查询收藏: %v", err)
	}
	if len(fs) != 3 {
		t.Fatalf("收藏行数 = %d, want 3（重复收藏不得新增行）: %+v", len(fs), fs)
	}
	want := []struct{ user, question uint }{{1, 100}, {1, 200}, {2, 100}}
	for i, w := range want {
		if fs[i].UserID != w.user || fs[i].QuestionID != w.question {
			t.Fatalf("第 %d 行归属错误: %+v", i, fs[i])
		}
	}
}

// TestFavoriteDeleteIdempotent 取消收藏幂等：取消不存在的收藏也返回成功
func TestFavoriteDeleteIdempotent(t *testing.T) {
	db := newFavoriteTestDB(t)
	repo := NewFavoriteRepository(db)

	// 未收藏直接取消：成功
	if err := repo.Delete(1, 100); err != nil {
		t.Fatalf("取消不存在的收藏应成功: %v", err)
	}
	if err := repo.Add(1, 100); err != nil {
		t.Fatalf("Add(1, 100): %v", err)
	}
	if err := repo.Delete(1, 100); err != nil {
		t.Fatalf("Delete(1, 100): %v", err)
	}
	// 已取消再取消：仍成功，且不影响他人
	if err := repo.Add(2, 100); err != nil {
		t.Fatalf("Add(2, 100): %v", err)
	}
	if err := repo.Delete(1, 100); err != nil {
		t.Fatalf("重复取消应成功: %v", err)
	}
	var total int64
	if err := db.Model(&model.Favorite{}).Count(&total).Error; err != nil {
		t.Fatalf("统计收藏: %v", err)
	}
	if total != 1 {
		t.Fatalf("重复取消后剩余收藏 = %d, want 1（仅用户 2）", total)
	}
}
