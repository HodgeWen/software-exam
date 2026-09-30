// Package seed 维护仓库内题库种子：JSON 为唯一数据源，Import 在启动时幂等导入。
package seed

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"software-exam/backend/internal/model"
)

//go:embed software_designer.json
var softwareDesigner []byte

type seedDoc struct {
	Subjects []seedSubject `json:"subjects"`
}

type seedSubject struct {
	Code     string        `json:"code"`
	Name     string        `json:"name"`
	Chapters []seedChapter `json:"chapters"`
	Papers   []seedPaper   `json:"papers"`
}

type seedChapter struct {
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	Questions []seedQuestion `json:"questions"`
}

type seedPaper struct {
	Code            string         `json:"code"`
	Name            string         `json:"name"`
	DurationMinutes int            `json:"duration_minutes"`
	Questions       []seedQuestion `json:"questions"`
}

type seedQuestion struct {
	No       int            `json:"no"`
	Type     string         `json:"type"`
	Stem     string         `json:"stem"`
	Options  []model.Option `json:"options"`
	Answer   []string       `json:"answer"`
	Analysis string         `json:"analysis"`
}

// Import 幂等导入全部种子：以稳定业务键（科目/章节/试卷编码 + 题号）判重，
// 已存在则更新内容、不存在则创建，重复执行不产生重复行。须在 AutoMigrate 之后调用。
func Import(db *gorm.DB) error {
	doc, err := load()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, s := range doc.Subjects {
			if err := importSubject(tx, s); err != nil {
				return err
			}
		}
		return nil
	})
}

func load() (seedDoc, error) {
	var doc seedDoc
	if err := json.Unmarshal(softwareDesigner, &doc); err != nil {
		return doc, fmt.Errorf("解析种子 JSON: %w", err)
	}
	return doc, nil
}

func importSubject(tx *gorm.DB, s seedSubject) error {
	subject := model.Subject{Code: s.Code}
	if err := tx.Where(&subject).Assign(model.Subject{Name: s.Name}).FirstOrCreate(&subject).Error; err != nil {
		return fmt.Errorf("导入科目 %s: %w", s.Code, err)
	}
	for i, c := range s.Chapters {
		chapter := model.Chapter{SubjectID: subject.ID, Code: c.Code}
		if err := tx.Where(&chapter).Assign(model.Chapter{Name: c.Name, Sort: i + 1}).FirstOrCreate(&chapter).Error; err != nil {
			return fmt.Errorf("导入章节 %s/%s: %w", s.Code, c.Code, err)
		}
		for _, q := range c.Questions {
			if _, err := importQuestion(tx, questionCode(s.Code, c.Code, q.No), subject.ID, &chapter.ID, q); err != nil {
				return err
			}
		}
	}
	for _, p := range s.Papers {
		if err := importPaper(tx, s.Code, subject.ID, p); err != nil {
			return err
		}
	}
	return nil
}

func importPaper(tx *gorm.DB, subjectCode string, subjectID uint, p seedPaper) error {
	paper := model.Paper{Code: p.Code}
	if err := tx.Where(&paper).Assign(model.Paper{
		SubjectID:       subjectID,
		Name:            p.Name,
		DurationMinutes: p.DurationMinutes,
	}).FirstOrCreate(&paper).Error; err != nil {
		return fmt.Errorf("导入试卷 %s: %w", p.Code, err)
	}

	links := make([]model.PaperQuestion, 0, len(p.Questions))
	for _, q := range p.Questions {
		question, err := importQuestion(tx, questionCode(subjectCode, p.Code, q.No), subjectID, nil, q)
		if err != nil {
			return err
		}
		links = append(links, model.PaperQuestion{PaperID: paper.ID, QuestionID: question.ID, No: q.No})
	}
	// 关联表内容完全由种子派生，整卷重建最简单且天然幂等
	if err := tx.Where("paper_id = ?", paper.ID).Delete(&model.PaperQuestion{}).Error; err != nil {
		return fmt.Errorf("重建试卷关联 %s: %w", p.Code, err)
	}
	if len(links) > 0 {
		if err := tx.Create(&links).Error; err != nil {
			return fmt.Errorf("写入试卷关联 %s: %w", p.Code, err)
		}
	}
	return nil
}

func importQuestion(tx *gorm.DB, code string, subjectID uint, chapterID *uint, q seedQuestion) (*model.Question, error) {
	question := model.Question{Code: code}
	if err := tx.Where(&question).Assign(model.Question{
		SubjectID: subjectID,
		ChapterID: chapterID,
		No:        q.No,
		Type:      q.Type,
		Stem:      q.Stem,
		Options:   q.Options,
		Answer:    q.Answer,
		Analysis:  q.Analysis,
	}).FirstOrCreate(&question).Error; err != nil {
		return nil, fmt.Errorf("导入题目 %s: %w", code, err)
	}
	return &question, nil
}

// questionCode 题目稳定业务键：科目编码 / 章节或试卷编码 / 三位题号
func questionCode(subjectCode, scopeCode string, no int) string {
	return fmt.Sprintf("%s/%s/q%03d", subjectCode, scopeCode, no)
}
