package repository

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

func newMistakeTestDB(t *testing.T) *gorm.DB {
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
	if err := db.AutoMigrate(&model.Mistake{}); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	return db
}

// TestUpsertOnWrongDeduplicates 错题判定：首次答错创建，重复答错仅累加错误次数并刷新最近答错时间，不产生重复行
func TestUpsertOnWrongDeduplicates(t *testing.T) {
	db := newMistakeTestDB(t)
	repo := NewMistakeRepository(db)

	first := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	second := first.Add(2 * time.Hour)

	for _, call := range []struct {
		user, question uint
		at             time.Time
	}{
		{1, 100, first},
		{1, 100, second}, // 同用户同题重复答错
		{2, 100, second}, // 其他用户互不影响
		{1, 200, second}, // 同用户不同题
	} {
		if err := repo.UpsertOnWrong(call.user, call.question, call.at); err != nil {
			t.Fatalf("UpsertOnWrong(%d, %d): %v", call.user, call.question, err)
		}
	}

	var ms []model.Mistake
	if err := db.Order("user_id, question_id").Find(&ms).Error; err != nil {
		t.Fatalf("查询错题: %v", err)
	}
	if len(ms) != 3 {
		t.Fatalf("错题行数 = %d, want 3（重复答错不得新增行）: %+v", len(ms), ms)
	}
	if ms[0].UserID != 1 || ms[0].QuestionID != 100 {
		t.Fatalf("第 1 行归属错误: %+v", ms[0])
	}
	if ms[0].WrongCount != 2 || !ms[0].LastWrongAt.Equal(second) {
		t.Fatalf("重复答错应累加错误次数并刷新最近答错时间: %+v", ms[0])
	}
	if ms[1].UserID != 1 || ms[1].QuestionID != 200 || ms[1].WrongCount != 1 {
		t.Fatalf("同用户不同题应独立成行: %+v", ms[1])
	}
	if ms[2].UserID != 2 || ms[2].QuestionID != 100 || ms[2].WrongCount != 1 {
		t.Fatalf("不同用户错题应隔离: %+v", ms[2])
	}
}
