package service

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

func newAnswerService(t *testing.T) (*AnswerService, *gorm.DB) {
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
	if err := db.AutoMigrate(&model.Question{}, &model.AnswerRecord{}, &model.Mistake{}); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	return NewAnswerService(
		repository.NewQuestionRepository(db),
		repository.NewAnswerRecordRepository(db),
		repository.NewMistakeRepository(db),
	), db
}

func seedJudgeQuestions(t *testing.T, db *gorm.DB) (single, multi *model.Question) {
	t.Helper()
	single = &model.Question{
		SubjectID: 1, Code: "P4-S01", No: 1, Type: model.QuestionTypeSingle,
		Stem: "单选题干", Answer: []string{"A"}, Analysis: "单选解析",
	}
	multi = &model.Question{
		SubjectID: 1, Code: "P4-M01", No: 2, Type: model.QuestionTypeMultiple,
		Stem: "多选题干", Answer: []string{"A", "C"}, Analysis: "多选解析",
	}
	for _, q := range []*model.Question{single, multi} {
		if err := db.Create(q).Error; err != nil {
			t.Fatalf("种子题目 %s: %v", q.Code, err)
		}
	}
	return single, multi
}

// TestJudge 判分单测：单选所选等于正确答案判对；多选所选集合与正确集合完全一致才判对
func TestJudge(t *testing.T) {
	single := &model.Question{Type: model.QuestionTypeSingle, Answer: []string{"A"}}
	multi := &model.Question{Type: model.QuestionTypeMultiple, Answer: []string{"A", "C"}}
	tests := []struct {
		name     string
		q        *model.Question
		selected []string
		want     bool
	}{
		{"单选所选等于正确答案", single, []string{"A"}, true},
		{"单选所选不同", single, []string{"B"}, false},
		{"单选未作答", single, nil, false},
		{"单选多选项", single, []string{"A", "B"}, false},
		{"多选完全一致", multi, []string{"A", "C"}, true},
		{"多选顺序无关", multi, []string{"C", "A"}, true},
		{"多选少选", multi, []string{"A"}, false},
		{"多选多选了额外项", multi, []string{"A", "B", "C"}, false},
		{"多选含错误项", multi, []string{"A", "B"}, false},
		{"多选重复选项", multi, []string{"A", "A"}, false},
		{"多选未作答", multi, nil, false},
		{"正确答案缺失视为无法判对", &model.Question{Type: model.QuestionTypeSingle}, []string{"A"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Judge(tt.q, tt.selected); got != tt.want {
				t.Fatalf("Judge(%s 型, %v) = %v, want %v", tt.q.Type, tt.selected, got, tt.want)
			}
		})
	}
}

// TestSubmitRecordsAndMistakes 提交沉淀：每次提交落一条 answer_record；答错入错题、
// 重刷答对仅更新记录不自动移除、业务数据按用户隔离、题目不存在报业务错误
func TestSubmitRecordsAndMistakes(t *testing.T) {
	s, db := newAnswerService(t)
	single, multi := seedJudgeQuestions(t, db)

	countRecords := func(user uint) (total, correct int64) {
		if err := db.Model(&model.AnswerRecord{}).Where("user_id = ?", user).Count(&total).Error; err != nil {
			t.Fatalf("统计答题记录: %v", err)
		}
		if err := db.Model(&model.AnswerRecord{}).Where("user_id = ? AND correct", user).Count(&correct).Error; err != nil {
			t.Fatalf("统计答对记录: %v", err)
		}
		return total, correct
	}
	mistakeOf := func(user, question uint) model.Mistake {
		var m model.Mistake
		if err := db.Where("user_id = ? AND question_id = ?", user, question).First(&m).Error; err != nil {
			t.Fatalf("查询错题(user=%d, question=%d): %v", user, question, err)
		}
		return m
	}

	// 答对：只落答题记录，不入错题
	q, correct, err := s.Submit(1, single.ID, []string{"A"})
	if err != nil || !correct || q.Analysis != "单选解析" {
		t.Fatalf("单选答对: correct=%v err=%v q=%+v", correct, err, q)
	}
	if total, right := countRecords(1); total != 1 || right != 1 {
		t.Fatalf("答对应落一条答对记录: total=%d correct=%d", total, right)
	}
	var mistakes int64
	db.Model(&model.Mistake{}).Where("user_id = ?", 1).Count(&mistakes)
	if mistakes != 0 {
		t.Fatalf("答对不应入错题: %d", mistakes)
	}

	// 多选少选判错：入错题，返回的题目带正确答案与解析供 handler 响应
	q, correct, err = s.Submit(1, multi.ID, []string{"A"})
	if err != nil || correct || len(q.Answer) != 2 {
		t.Fatalf("多选少选应判错: correct=%v err=%v answer=%v", correct, err, q.Answer)
	}
	if m := mistakeOf(1, multi.ID); m.WrongCount != 1 {
		t.Fatalf("首次答错 wrong_count 应为 1: %+v", m)
	}

	// 重刷答对：仅更新记录，不自动移除错题
	if _, correct, err = s.Submit(1, multi.ID, []string{"C", "A"}); err != nil || !correct {
		t.Fatalf("多选顺序无关应判对: correct=%v err=%v", correct, err)
	}
	if m := mistakeOf(1, multi.ID); m.WrongCount != 1 {
		t.Fatalf("重刷答对不应移除或累错题: %+v", m)
	}
	if total, right := countRecords(1); total != 3 || right != 2 {
		t.Fatalf("每次提交都应落一条记录: total=%d correct=%d", total, right)
	}

	// 业务数据按用户隔离：用户 2 答错不影响用户 1
	if _, _, err = s.Submit(2, multi.ID, []string{"B"}); err != nil {
		t.Fatalf("用户 2 提交: %v", err)
	}
	if m := mistakeOf(1, multi.ID); m.WrongCount != 1 {
		t.Fatalf("用户 2 答错不应影响用户 1 错题: %+v", m)
	}
	if m := mistakeOf(2, multi.ID); m.WrongCount != 1 {
		t.Fatalf("用户 2 应有自己的错题行: %+v", m)
	}

	// 题目不存在
	if _, _, err = s.Submit(1, 999999, []string{"A"}); !errors.Is(err, ErrQuestionNotFound) {
		t.Fatalf("题目不存在未返回 ErrQuestionNotFound: %v", err)
	}
}
