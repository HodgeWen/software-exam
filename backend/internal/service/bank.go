package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
	"software-exam/backend/internal/repository"
)

var (
	ErrSubjectNotFound = errors.New("科目不存在")
	ErrChapterNotFound = errors.New("章节不存在")
	ErrPaperNotFound   = errors.New("试卷不存在")
)

// SubjectDetail 科目详情聚合：科目 + 章节列表 + 该科目试卷列表
type SubjectDetail struct {
	Subject  model.Subject
	Chapters []model.Chapter
	Papers   []model.Paper
}

// BankService 题库取题业务：科目/章节/试卷查询与顺序、随机两种取题方式
type BankService struct {
	repo *repository.BankRepository
}

func NewBankService(repo *repository.BankRepository) *BankService {
	return &BankService{repo: repo}
}

func (s *BankService) Subjects() ([]model.Subject, error) {
	subjects, err := s.repo.ListSubjects()
	if err != nil {
		return nil, fmt.Errorf("查询科目列表: %w", err)
	}
	return subjects, nil
}

func (s *BankService) SubjectDetail(subjectID uint) (*SubjectDetail, error) {
	subject, err := s.findSubject(subjectID)
	if err != nil {
		return nil, err
	}
	chapters, err := s.repo.ListChapters(subjectID)
	if err != nil {
		return nil, fmt.Errorf("查询章节列表: %w", err)
	}
	papers, err := s.repo.ListPapersBySubject(subjectID)
	if err != nil {
		return nil, fmt.Errorf("查询试卷列表: %w", err)
	}
	return &SubjectDetail{
		Subject:  *subject,
		Chapters: chapters,
		Papers:   papers,
	}, nil
}

// ChapterQuestions 顺序取题：校验科目与章节归属后，按题目稳定顺序（ID）返回章节全部题目
func (s *BankService) ChapterQuestions(subjectID, chapterID uint) ([]model.Question, error) {
	if err := s.ensureSubject(subjectID); err != nil {
		return nil, err
	}
	if err := s.ensureChapter(subjectID, chapterID); err != nil {
		return nil, err
	}
	questions, err := s.repo.ListQuestionsByChapter(chapterID)
	if err != nil {
		return nil, fmt.Errorf("查询章节题目: %w", err)
	}
	return questions, nil
}

// RandomQuestions 随机取题：按科目（可选限定章节）取全部题目后打乱返回
func (s *BankService) RandomQuestions(subjectID uint, chapterID *uint) ([]model.Question, error) {
	if err := s.ensureSubject(subjectID); err != nil {
		return nil, err
	}
	var questions []model.Question
	var err error
	if chapterID != nil {
		if err := s.ensureChapter(subjectID, *chapterID); err != nil {
			return nil, err
		}
		questions, err = s.repo.ListQuestionsByChapter(*chapterID)
	} else {
		questions, err = s.repo.ListQuestionsBySubject(subjectID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询随机题源: %w", err)
	}
	return shuffleQuestions(questions, nil), nil
}

func (s *BankService) Papers() ([]model.Paper, error) {
	papers, err := s.repo.ListPapers()
	if err != nil {
		return nil, fmt.Errorf("查询试卷列表: %w", err)
	}
	return papers, nil
}

// PaperQuestions 整卷题目，按卷内题号升序；题目本体由 handler 裁剪为不含答案与解析的视图
func (s *BankService) PaperQuestions(paperID uint) (*model.Paper, []repository.PaperQuestionItem, error) {
	paper, err := s.repo.FindPaper(paperID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrPaperNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("查询试卷: %w", err)
	}
	items, err := s.repo.ListPaperQuestions(paperID)
	if err != nil {
		return nil, nil, fmt.Errorf("查询试卷题目: %w", err)
	}
	return paper, items, nil
}

// ensureSubject 校验科目存在，不存在返回 ErrSubjectNotFound
func (s *BankService) ensureSubject(subjectID uint) error {
	_, err := s.findSubject(subjectID)
	return err
}

// findSubject 查科目并归一错误：不存在转 ErrSubjectNotFound，其余包上下文
func (s *BankService) findSubject(subjectID uint) (*model.Subject, error) {
	subject, err := s.repo.FindSubject(subjectID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSubjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询科目: %w", err)
	}
	return subject, nil
}

// ensureChapter 校验章节存在且属于指定科目
func (s *BankService) ensureChapter(subjectID, chapterID uint) error {
	_, err := s.repo.FindChapter(subjectID, chapterID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrChapterNotFound
	}
	if err != nil {
		return fmt.Errorf("查询章节: %w", err)
	}
	return nil
}
