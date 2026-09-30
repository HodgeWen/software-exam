package service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

var ErrQuestionNotFound = errors.New("题目不存在")

// Judge 判分：所选集合与正确答案集合完全一致才判对——单选正确答案仅 1 个标号，
// 集合一致等价于所选等于正确答案；多选要求不重不漏（少选/多选/重复选均判错）
func Judge(q *model.Question, selected []string) bool {
	if len(q.Answer) == 0 || len(selected) != len(q.Answer) {
		return false
	}
	remain := make(map[string]int, len(q.Answer))
	for _, a := range q.Answer {
		remain[a]++
	}
	for _, s := range selected {
		if remain[s] == 0 {
			return false
		}
		remain[s]--
	}
	return true
}

// AnswerService 单题提交判分与记录沉淀；考试交卷（P6）复用 Judge 与 Record 逐题判分落库
type AnswerService struct {
	questions *repository.QuestionRepository
	records   *repository.AnswerRecordRepository
	mistakes  *repository.MistakeRepository
}

func NewAnswerService(
	questions *repository.QuestionRepository,
	records *repository.AnswerRecordRepository,
	mistakes *repository.MistakeRepository,
) *AnswerService {
	return &AnswerService{questions: questions, records: records, mistakes: mistakes}
}

// Submit 单题提交：取题判分并沉淀记录，返回判过的题目与对错，供 handler 组装对/错、正确答案与解析
func (s *AnswerService) Submit(userID, questionID uint, selected []string) (*model.Question, bool, error) {
	q, err := s.questions.FindByID(questionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, ErrQuestionNotFound
	}
	if err != nil {
		return nil, false, fmt.Errorf("查询题目: %w", err)
	}
	correct := Judge(q, selected)
	if err := s.Record(userID, q, selected, correct); err != nil {
		return nil, false, err
	}
	return q, correct, nil
}

// Record 沉淀一次作答：每次提交落一条 answer_record；答错同时 upsert 错题本。
// 重刷/考试答错与练习走同一入口；答对仅更新记录，不移除已有错题
func (s *AnswerService) Record(userID uint, q *model.Question, selected []string, correct bool) error {
	rec := &model.AnswerRecord{UserID: userID, QuestionID: q.ID, Selected: selected, Correct: correct}
	if err := s.records.Create(rec); err != nil {
		return fmt.Errorf("落答题记录: %w", err)
	}
	if !correct {
		if err := s.mistakes.UpsertOnWrong(userID, q.ID, time.Now()); err != nil {
			return fmt.Errorf("更新错题: %w", err)
		}
	}
	return nil
}
