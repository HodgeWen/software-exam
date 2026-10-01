package seed

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

// newTestDB 内存 SQLite 建出全部题库表；内存库每个连接独立，限单连接避免「no such table」
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("打开内存 SQLite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("取底层 sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&model.Subject{}, &model.Chapter{}, &model.Question{},
		&model.Paper{}, &model.PaperQuestion{},
	); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	return db
}

func count(t *testing.T, db *gorm.DB, dst any) int64 {
	t.Helper()
	var n int64
	if err := db.Model(dst).Count(&n).Error; err != nil {
		t.Fatalf("计数 %T: %v", dst, err)
	}
	return n
}

// TestImportIdempotent 同一份种子连导两次：行数不变，且被外部改掉的内容会被种子更新回来
func TestImportIdempotent(t *testing.T) {
	db := newTestDB(t)
	if err := Import(db); err != nil {
		t.Fatalf("首次导入: %v", err)
	}
	before := map[string]int64{
		"subjects":        count(t, db, &model.Subject{}),
		"chapters":        count(t, db, &model.Chapter{}),
		"questions":       count(t, db, &model.Question{}),
		"papers":          count(t, db, &model.Paper{}),
		"paper_questions": count(t, db, &model.PaperQuestion{}),
	}

	// 篡改一题内容，验证重复导入不丢更新
	var chapterID uint
	db.Model(&model.Chapter{}).Select("id").Where("code = ?", "computer-system").Scan(&chapterID)
	if err := db.Model(&model.Question{}).Where("chapter_id = ? AND no = ?", chapterID, 1).
		Updates(map[string]any{"stem": "被篡改", "analysis": ""}).Error; err != nil {
		t.Fatalf("篡改题目: %v", err)
	}

	if err := Import(db); err != nil {
		t.Fatalf("重复导入: %v", err)
	}

	for name, want := range before {
		var got int64
		switch name {
		case "subjects":
			got = count(t, db, &model.Subject{})
		case "chapters":
			got = count(t, db, &model.Chapter{})
		case "questions":
			got = count(t, db, &model.Question{})
		case "papers":
			got = count(t, db, &model.Paper{})
		case "paper_questions":
			got = count(t, db, &model.PaperQuestion{})
		}
		if got != want {
			t.Errorf("重复导入后 %s 行数变化: 第一次 %d, 第二次 %d", name, want, got)
		}
	}

	var q model.Question
	if err := db.Where("chapter_id = ? AND no = ?", chapterID, 1).First(&q).Error; err != nil {
		t.Fatalf("回读被篡改题目: %v", err)
	}
	if q.Stem == "被篡改" || q.Analysis == "" {
		t.Errorf("重复导入未恢复被篡改内容: stem=%q, analysis=%q", q.Stem, q.Analysis)
	}
}

// TestSeedContent 校验种子内容完整性：各科目的章节、练习题、真题数量下限，
// 并逐题校验答案合法（在选项内、单选恰 1 个、多选 ≥2 个）且有解析、题型覆盖单选与多选
func TestSeedContent(t *testing.T) {
	db := newTestDB(t)
	if err := Import(db); err != nil {
		t.Fatalf("导入: %v", err)
	}

	specs := []struct {
		code                                string
		minChapters, minPractice, minPapers int64
	}{
		{"soft-designer", 4, 40, 1},
		{"sys-architect", 8, 80, 2},
		{"it-pm", 7, 70, 2},
	}
	for _, spec := range specs {
		t.Run(spec.code, func(t *testing.T) {
			checkSubjectContent(t, db, spec.code, spec.minChapters, spec.minPractice, spec.minPapers)
		})
	}
}

func checkSubjectContent(t *testing.T, db *gorm.DB, code string, minChapters, minPractice, minPapers int64) {
	t.Helper()

	var subject model.Subject
	if err := db.Where("code = ?", code).First(&subject).Error; err != nil {
		t.Fatalf("查科目 %s: %v", code, err)
	}

	var chapters int64
	if err := db.Model(&model.Chapter{}).Where("subject_id = ?", subject.ID).Count(&chapters).Error; err != nil {
		t.Fatalf("统计章节数: %v", err)
	}
	if chapters < minChapters {
		t.Errorf("章节数 %d < %d", chapters, minChapters)
	}

	var practiceCount int64
	if err := db.Model(&model.Question{}).Where("subject_id = ? AND chapter_id IS NOT NULL", subject.ID).
		Count(&practiceCount).Error; err != nil {
		t.Fatalf("统计章节练习题数: %v", err)
	}
	if practiceCount < minPractice {
		t.Errorf("章节练习题数 %d < %d", practiceCount, minPractice)
	}

	var papers []model.Paper
	if err := db.Where("subject_id = ?", subject.ID).Find(&papers).Error; err != nil {
		t.Fatalf("查试卷: %v", err)
	}
	if int64(len(papers)) < minPapers {
		t.Fatalf("真题试卷数 %d < %d", len(papers), minPapers)
	}
	for _, p := range papers {
		var n int64
		if err := db.Model(&model.PaperQuestion{}).Where("paper_id = ?", p.ID).Count(&n).Error; err != nil {
			t.Fatalf("统计试卷 %s 题数: %v", p.Code, err)
		}
		if n < 20 {
			t.Errorf("试卷 %s 题数 %d < 20", p.Code, n)
		}
		if p.DurationMinutes <= 0 {
			t.Errorf("试卷 %s 时长 %d 非法", p.Code, p.DurationMinutes)
		}
	}

	var questions []model.Question
	if err := db.Where("subject_id = ?", subject.ID).Order("code").Find(&questions).Error; err != nil {
		t.Fatalf("查全部题目: %v", err)
	}
	typeCount := map[string]int{}
	for _, q := range questions {
		keys := map[string]bool{}
		for _, o := range q.Options {
			keys[o.Key] = true
		}
		if len(q.Options) < 2 {
			t.Errorf("%s 选项数 %d < 2", q.Code, len(q.Options))
		}
		if q.Analysis == "" {
			t.Errorf("%s 缺少解析", q.Code)
		}
		if len(q.Answer) == 0 {
			t.Errorf("%s 缺少正确答案", q.Code)
		}
		for _, a := range q.Answer {
			if !keys[a] {
				t.Errorf("%s 答案 %s 不在选项内", q.Code, a)
			}
		}
		switch q.Type {
		case model.QuestionTypeSingle:
			if len(q.Answer) != 1 {
				t.Errorf("%s 单选答案数 %d != 1", q.Code, len(q.Answer))
			}
		case model.QuestionTypeMultiple:
			if len(q.Answer) < 2 {
				t.Errorf("%s 多选答案数 %d < 2", q.Code, len(q.Answer))
			}
		default:
			t.Errorf("%s 非法题型 %q", q.Code, q.Type)
		}
		typeCount[q.Type]++
	}
	if typeCount[model.QuestionTypeSingle] == 0 || typeCount[model.QuestionTypeMultiple] == 0 {
		t.Errorf("题型覆盖不全: %v", typeCount)
	}
}
