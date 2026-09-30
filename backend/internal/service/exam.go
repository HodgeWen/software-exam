package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

var (
	ErrExamNotFound         = errors.New("考试记录不存在")
	ErrExamAlreadySubmitted = errors.New("该考试已交卷")
)

// ExamQuestionResult 交卷后逐题判分明细：所选、对错与题目本体（含正确答案/解析，供 handler 组装）
type ExamQuestionResult struct {
	Question model.Question
	No       int
	Selected []string
	Correct  bool
}

// ExamSubmitResult 交卷统一评分结果：考试记录得分快照 + 逐题判分明细
type ExamSubmitResult struct {
	Exam    *model.Exam
	Results []ExamQuestionResult
}

// ExamService 真题模拟考试：开始生成进行中记录，交卷一次性统一评分；
// 逐题判分与沉淀复用 P4 的 Judge 与 AnswerService.Record
type ExamService struct {
	bank    *repository.BankRepository
	exams   *repository.ExamRepository
	answers *AnswerService
}

func NewExamService(
	bank *repository.BankRepository,
	exams *repository.ExamRepository,
	answers *AnswerService,
) *ExamService {
	return &ExamService{bank: bank, exams: exams, answers: answers}
}

// Start 开始考试：生成进行中的考试记录并返回试卷与整卷题目（题目由 handler 裁剪为不含答案与解析的视图）
func (s *ExamService) Start(userID, paperID uint) (*model.Exam, *model.Paper, []repository.PaperQuestionItem, error) {
	paper, err := s.bank.FindPaper(paperID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil, ErrPaperNotFound
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("查询试卷: %w", err)
	}
	items, err := s.bank.ListPaperQuestions(paperID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("查询试卷题目: %w", err)
	}
	exam := &model.Exam{
		UserID:    userID,
		PaperID:   paperID,
		Status:    model.ExamStatusInProgress,
		StartedAt: time.Now(),
	}
	if err := s.exams.Create(exam); err != nil {
		return nil, nil, nil, fmt.Errorf("创建考试记录: %w", err)
	}
	return exam, paper, items, nil
}

// Submit 交卷：以试卷题目为基准一次性统一评分，未作答视为空所选判错；
// 逐题复用 Judge 判分、Record 落 answer_record 与错题本，最后把考试记录流转为已交卷并沉淀得分快照。
// 已交卷的考试记录返回 ErrExamAlreadySubmitted，不产生新的答题记录
func (s *ExamService) Submit(userID, examID uint, answers map[uint][]string) (*ExamSubmitResult, error) {
	exam, err := s.exams.FindByUser(userID, examID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrExamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询考试记录: %w", err)
	}
	if exam.Status != model.ExamStatusInProgress {
		return nil, ErrExamAlreadySubmitted
	}
	items, err := s.bank.ListPaperQuestions(exam.PaperID)
	if err != nil {
		return nil, fmt.Errorf("查询试卷题目: %w", err)
	}

	result := &ExamSubmitResult{Exam: exam, Results: make([]ExamQuestionResult, 0, len(items))}
	for _, item := range items {
		q := item.Question
		selected := answers[q.ID]
		correct := Judge(&q, selected)
		if err := s.answers.Record(userID, &q, selected, correct); err != nil {
			return nil, err
		}
		if correct {
			exam.CorrectCount++
		}
		result.Results = append(result.Results, ExamQuestionResult{
			Question: q, No: item.No, Selected: selected, Correct: correct,
		})
	}
	exam.TotalCount = len(items)
	if exam.TotalCount > 0 {
		exam.Accuracy = float64(exam.CorrectCount) / float64(exam.TotalCount)
	}
	now := time.Now()
	exam.Status = model.ExamStatusSubmitted
	exam.SubmittedAt = &now
	if err := s.exams.Save(exam); err != nil {
		return nil, fmt.Errorf("更新考试记录: %w", err)
	}
	return result, nil
}
