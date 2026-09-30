package repository

import (
	"software-exam/backend/internal/model"

	"gorm.io/gorm"
)

// BankRepository 题库只读查询：科目/章节/试卷/题目，供取题接口使用
type BankRepository struct {
	db *gorm.DB
}

func NewBankRepository(db *gorm.DB) *BankRepository {
	return &BankRepository{db: db}
}

func (r *BankRepository) FindSubject(id uint) (*model.Subject, error) {
	var subject model.Subject
	if err := r.db.First(&subject, id).Error; err != nil {
		return nil, err
	}
	return &subject, nil
}

func (r *BankRepository) ListSubjects() ([]model.Subject, error) {
	subjects := make([]model.Subject, 0)
	err := r.db.Order("id").Find(&subjects).Error
	return subjects, err
}

// FindChapter 定位科目下的章节；(subject_id, id) 双条件，章节跨科目引用即视为不存在
func (r *BankRepository) FindChapter(subjectID, chapterID uint) (*model.Chapter, error) {
	var chapter model.Chapter
	err := r.db.Where("subject_id = ? AND id = ?", subjectID, chapterID).First(&chapter).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

func (r *BankRepository) ListChapters(subjectID uint) ([]model.Chapter, error) {
	chapters := make([]model.Chapter, 0)
	err := r.db.Where("subject_id = ?", subjectID).Order("sort").Find(&chapters).Error
	return chapters, err
}

func (r *BankRepository) FindPaper(id uint) (*model.Paper, error) {
	var paper model.Paper
	if err := r.db.First(&paper, id).Error; err != nil {
		return nil, err
	}
	return &paper, nil
}

func (r *BankRepository) ListPapers() ([]model.Paper, error) {
	papers := make([]model.Paper, 0)
	err := r.db.Order("id").Find(&papers).Error
	return papers, err
}

func (r *BankRepository) ListPapersBySubject(subjectID uint) ([]model.Paper, error) {
	papers := make([]model.Paper, 0)
	err := r.db.Where("subject_id = ?", subjectID).Order("id").Find(&papers).Error
	return papers, err
}

// ListQuestionsByChapter 章节练习题，按题目 ID 升序（即种子导入的稳定顺序）
func (r *BankRepository) ListQuestionsByChapter(chapterID uint) ([]model.Question, error) {
	questions := make([]model.Question, 0)
	err := r.db.Where("chapter_id = ?", chapterID).Order("id").Find(&questions).Error
	return questions, err
}

// ListQuestionsBySubject 科目全部题目（含试卷独占题），作为随机练习的题源
func (r *BankRepository) ListQuestionsBySubject(subjectID uint) ([]model.Question, error) {
	questions := make([]model.Question, 0)
	err := r.db.Where("subject_id = ?", subjectID).Order("id").Find(&questions).Error
	return questions, err
}

// PaperQuestionItem 试卷题目项：No 为卷内题号，Question 为题目本体
type PaperQuestionItem struct {
	No       int
	Question model.Question
}

// ListPaperQuestions 整卷题目，按卷内题号升序；两段查询避免 join 结果与模型扫描对不齐
func (r *BankRepository) ListPaperQuestions(paperID uint) ([]PaperQuestionItem, error) {
	links := make([]model.PaperQuestion, 0)
	if err := r.db.Where("paper_id = ?", paperID).Order("no").Find(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []PaperQuestionItem{}, nil
	}
	ids := make([]uint, len(links))
	for i, link := range links {
		ids[i] = link.QuestionID
	}
	questions := make([]model.Question, 0, len(ids))
	if err := r.db.Where("id IN ?", ids).Find(&questions).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint]model.Question, len(questions))
	for _, q := range questions {
		byID[q.ID] = q
	}
	items := make([]PaperQuestionItem, 0, len(links))
	for _, link := range links {
		items = append(items, PaperQuestionItem{No: link.No, Question: byID[link.QuestionID]})
	}
	return items, nil
}
