package service

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

// newExamService 内存库装配考试服务（含 P4 判分/沉淀依赖），表结构与启动迁移一致
func newExamService(t *testing.T) (*ExamService, *gorm.DB) {
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
	if err := db.AutoMigrate(
		&model.Subject{}, &model.Question{},
		&model.Paper{}, &model.PaperQuestion{},
		&model.Exam{}, &model.AnswerRecord{}, &model.Mistake{},
	); err != nil {
		t.Fatalf("迁移: %v", err)
	}
	return NewExamService(
		repository.NewBankRepository(db),
		repository.NewExamRepository(db),
		NewAnswerService(
			repository.NewQuestionRepository(db),
			repository.NewAnswerRecordRepository(db),
			repository.NewMistakeRepository(db),
		),
	), db
}

// examFixtures 一套 4 题试卷：单选×2、多选×2，覆盖答对/少选/多选/未作答四种结局
type examFixtures struct {
	paper                              model.Paper
	singleA, multiAC, multiBD, singleC model.Question
}

func seedExamFixtures(t *testing.T, db *gorm.DB) *examFixtures {
	t.Helper()
	subj := model.Subject{Code: "p6-subj", Name: "考试科目"}
	if err := db.Create(&subj).Error; err != nil {
		t.Fatalf("种子科目: %v", err)
	}
	f := &examFixtures{
		paper: model.Paper{SubjectID: subj.ID, Code: "p6-paper", Name: "模拟卷", DurationMinutes: 60},
	}
	if err := db.Create(&f.paper).Error; err != nil {
		t.Fatalf("种子试卷: %v", err)
	}
	f.singleA = model.Question{SubjectID: subj.ID, Code: "P6-S01", No: 1,
		Type: model.QuestionTypeSingle, Stem: "单选一", Answer: []string{"A"}, Analysis: "解析一"}
	f.multiAC = model.Question{SubjectID: subj.ID, Code: "P6-M01", No: 2,
		Type: model.QuestionTypeMultiple, Stem: "多选一", Answer: []string{"A", "C"}, Analysis: "解析二"}
	f.multiBD = model.Question{SubjectID: subj.ID, Code: "P6-M02", No: 3,
		Type: model.QuestionTypeMultiple, Stem: "多选二", Answer: []string{"B", "D"}, Analysis: "解析三"}
	f.singleC = model.Question{SubjectID: subj.ID, Code: "P6-S02", No: 4,
		Type: model.QuestionTypeSingle, Stem: "单选二", Answer: []string{"C"}, Analysis: "解析四"}
	for i, q := range []*model.Question{&f.singleA, &f.multiAC, &f.multiBD, &f.singleC} {
		if err := db.Create(q).Error; err != nil {
			t.Fatalf("种子题目 %s: %v", q.Code, err)
		}
		link := model.PaperQuestion{PaperID: f.paper.ID, QuestionID: q.ID, No: i + 1}
		if err := db.Create(&link).Error; err != nil {
			t.Fatalf("关联试卷题目 %s: %v", q.Code, err)
		}
	}
	return f
}

// TestExamStart 开始考试：生成进行中记录并按卷内题号返回整卷题目
func TestExamStart(t *testing.T) {
	s, db := newExamService(t)
	f := seedExamFixtures(t, db)

	exam, paper, items, err := s.Start(1, f.paper.ID)
	if err != nil {
		t.Fatalf("开始考试: %v", err)
	}
	if exam.Status != model.ExamStatusInProgress || exam.UserID != 1 || exam.StartedAt.IsZero() {
		t.Fatalf("考试记录应属于该用户且进行中: %+v", exam)
	}
	if paper.DurationMinutes != 60 || len(items) != 4 {
		t.Fatalf("应返回试卷时长与整卷题目: paper=%+v items=%d", paper, len(items))
	}
	for i, item := range items {
		if item.No != i+1 {
			t.Fatalf("整卷题目应按卷内题号升序: items[%d].No = %d", i, item.No)
		}
	}
	if _, _, _, err := s.Start(1, 999999); !errors.Is(err, ErrPaperNotFound) {
		t.Fatalf("未知试卷应返回 ErrPaperNotFound: %v", err)
	}
}

// TestExamSubmitGrading 交卷统一评分：多选少选/多选判错、未作答判错，
// 每题落 answer_record、答错入错题本，考试记录沉淀得分快照并流转为已交卷
func TestExamSubmitGrading(t *testing.T) {
	s, db := newExamService(t)
	f := seedExamFixtures(t, db)
	exam, _, _, err := s.Start(1, f.paper.ID)
	if err != nil {
		t.Fatalf("开始考试: %v", err)
	}

	// 单选一答对；多选一少选；多选二多选了额外项；单选二未提交（视为未作答）
	answers := map[uint][]string{
		f.singleA.ID: {"A"},
		f.multiAC.ID: {"A"},
		f.multiBD.ID: {"A", "B", "D"},
	}
	result, err := s.Submit(1, exam.ID, answers)
	if err != nil {
		t.Fatalf("交卷: %v", err)
	}
	if len(result.Results) != 4 {
		t.Fatalf("逐题结果应为 4 条: %d", len(result.Results))
	}
	wantCorrect := map[int]bool{1: true, 2: false, 3: false, 4: false}
	for _, r := range result.Results {
		if r.Correct != wantCorrect[r.No] {
			t.Fatalf("第 %d 题判分 = %v, want %v", r.No, r.Correct, wantCorrect[r.No])
		}
		if len(r.Question.Answer) == 0 || r.Question.Analysis == "" {
			t.Fatalf("逐题结果应带正确答案与解析: %+v", r)
		}
	}
	if result.Results[3].Selected != nil {
		t.Fatalf("未作答的题所选应为空: %+v", result.Results[3])
	}
	e := result.Exam
	if e.Status != model.ExamStatusSubmitted || e.SubmittedAt == nil {
		t.Fatalf("交卷后状态应流转为已交卷: %+v", e)
	}
	if e.CorrectCount != 1 || e.TotalCount != 4 || e.Accuracy != 0.25 {
		t.Fatalf("得分快照 correct=%d total=%d accuracy=%v, want 1/4/0.25",
			e.CorrectCount, e.TotalCount, e.Accuracy)
	}

	// 逐题沉淀：4 条答题记录（1 对 3 错），3 道错题入错题本
	var records, correctRecords int64
	if err := db.Model(&model.AnswerRecord{}).Where("user_id = ?", 1).Count(&records).Error; err != nil || records != 4 {
		t.Fatalf("答题记录 = %d (err %v), want 4", records, err)
	}
	if err := db.Model(&model.AnswerRecord{}).Where("user_id = ? AND correct", 1).
		Count(&correctRecords).Error; err != nil || correctRecords != 1 {
		t.Fatalf("答对记录 = %d (err %v), want 1", correctRecords, err)
	}
	var wrongQuestions []uint
	if err := db.Model(&model.Mistake{}).Where("user_id = ?", 1).
		Order("question_id").Pluck("question_id", &wrongQuestions).Error; err != nil {
		t.Fatalf("查询错题: %v", err)
	}
	wantWrong := []uint{f.multiAC.ID, f.multiBD.ID, f.singleC.ID}
	if len(wrongQuestions) != 3 || wrongQuestions[0] != wantWrong[0] ||
		wrongQuestions[1] != wantWrong[1] || wrongQuestions[2] != wantWrong[2] {
		t.Fatalf("答错应入错题本 %v, got %v", wantWrong, wrongQuestions)
	}

	// 考试记录落库：已交卷且带得分快照
	var stored model.Exam
	if err := db.First(&stored, exam.ID).Error; err != nil {
		t.Fatalf("查询考试记录: %v", err)
	}
	if stored.Status != model.ExamStatusSubmitted || stored.CorrectCount != 1 || stored.TotalCount != 4 {
		t.Fatalf("落库考试记录不符: %+v", stored)
	}
}

// TestExamSubmitRejected 重复交卷返回业务错误且不再产生答题记录；
// 他人/不存在的考试记录一律视为不存在
func TestExamSubmitRejected(t *testing.T) {
	s, db := newExamService(t)
	f := seedExamFixtures(t, db)
	exam, _, _, err := s.Start(1, f.paper.ID)
	if err != nil {
		t.Fatalf("开始考试: %v", err)
	}
	allCorrect := map[uint][]string{
		f.singleA.ID: {"A"},
		f.multiAC.ID: {"C", "A"},
		f.multiBD.ID: {"B", "D"},
		f.singleC.ID: {"C"},
	}
	if _, err := s.Submit(1, exam.ID, allCorrect); err != nil {
		t.Fatalf("首次交卷: %v", err)
	}
	if _, err := s.Submit(1, exam.ID, allCorrect); !errors.Is(err, ErrExamAlreadySubmitted) {
		t.Fatalf("重复交卷应返回 ErrExamAlreadySubmitted: %v", err)
	}
	if _, err := s.Submit(2, exam.ID, allCorrect); !errors.Is(err, ErrExamNotFound) {
		t.Fatalf("他人考试记录应视为不存在: %v", err)
	}
	if _, err := s.Submit(1, 999999, nil); !errors.Is(err, ErrExamNotFound) {
		t.Fatalf("不存在考试记录应返回 ErrExamNotFound: %v", err)
	}
	var records int64
	if err := db.Model(&model.AnswerRecord{}).Where("user_id = ?", 1).Count(&records).Error; err != nil || records != 4 {
		t.Fatalf("重复交卷不应新增答题记录 = %d (err %v), want 4", records, err)
	}
}
