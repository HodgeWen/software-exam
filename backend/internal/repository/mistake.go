package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"software-exam/backend/internal/model"
)

// MistakeWithQuestion 错题本列表行：错题记录 + 对应题目 + 归属科目/知识点名称（按用户隔离）
type MistakeWithQuestion struct {
	Mistake     model.Mistake
	Question    model.Question
	SubjectName string
	ChapterName string
}

type MistakeRepository struct {
	db *gorm.DB
}

func NewMistakeRepository(db *gorm.DB) *MistakeRepository {
	return &MistakeRepository{db: db}
}

// UpsertOnWrong 答错入库：单条 upsert 保证并发下也不产生重复行——
// 首次答错创建（wrong_count=1），重复答错仅累加错误次数并刷新最近答错时间
func (r *MistakeRepository) UpsertOnWrong(userID, questionID uint, at time.Time) error {
	m := model.Mistake{UserID: userID, QuestionID: questionID, WrongCount: 1, LastWrongAt: at}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"wrong_count":   gorm.Expr("wrong_count + 1"),
			"last_wrong_at": at,
			"updated_at":    at,
		}),
	}).Create(&m).Error
}

// List 错题本分页列表；subjectID > 0 时按题目所属科目过滤，按最近答错时间倒序。
// 过滤条件涉及题目表，Count 与取页分别执行，避免复用同一语句的子句残留
func (r *MistakeRepository) List(userID, subjectID uint, page, pageSize int) ([]MistakeWithQuestion, int64, error) {
	filtered := func() *gorm.DB {
		q := r.db.Model(&model.Mistake{}).Where("mistakes.user_id = ?", userID)
		if subjectID > 0 {
			q = q.Joins("JOIN questions ON questions.id = mistakes.question_id").
				Where("questions.subject_id = ?", subjectID)
		}
		return q
	}

	var total int64
	if err := filtered().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ms []model.Mistake
	if err := filtered().Select("mistakes.*").
		Order("mistakes.last_wrong_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&ms).Error; err != nil {
		return nil, 0, err
	}
	if len(ms) == 0 {
		return nil, total, nil
	}

	// 第二跳批量取题目，按错题顺序拼回列表行
	ids := make([]uint, len(ms))
	for i, m := range ms {
		ids[i] = m.QuestionID
	}
	var qs []model.Question
	if err := r.db.Find(&qs, ids).Error; err != nil {
		return nil, 0, err
	}
	byID := make(map[uint]model.Question, len(qs))
	for _, q := range qs {
		byID[q.ID] = q
	}
	rows := make([]MistakeWithQuestion, 0, len(ms))
	for _, m := range ms {
		if q, ok := byID[m.QuestionID]; ok {
			rows = append(rows, MistakeWithQuestion{Mistake: m, Question: q})
		}
	}
	if err := r.fillScopeNames(rows); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// fillScopeNames 第三跳批量补齐每行题目的科目与知识点（章节）名称；真题题无章节则为空
func (r *MistakeRepository) fillScopeNames(rows []MistakeWithQuestion) error {
	chapterIDs := make([]uint, 0, len(rows))
	subjectIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.Question.ChapterID != nil {
			chapterIDs = append(chapterIDs, *row.Question.ChapterID)
		}
		subjectIDs = append(subjectIDs, row.Question.SubjectID)
	}

	chapterNames := map[uint]string{}
	if len(chapterIDs) > 0 {
		var chapters []model.Chapter
		if err := r.db.Find(&chapters, chapterIDs).Error; err != nil {
			return err
		}
		for _, c := range chapters {
			chapterNames[c.ID] = c.Name
		}
	}
	subjectNames := map[uint]string{}
	if len(subjectIDs) > 0 {
		var subjects []model.Subject
		if err := r.db.Find(&subjects, subjectIDs).Error; err != nil {
			return err
		}
		for _, s := range subjects {
			subjectNames[s.ID] = s.Name
		}
	}
	for i := range rows {
		rows[i].SubjectName = subjectNames[rows[i].Question.SubjectID]
		if rows[i].Question.ChapterID != nil {
			rows[i].ChapterName = chapterNames[*rows[i].Question.ChapterID]
		}
	}
	return nil
}

// ListQuestions 错题重刷题源：以当前用户错题为题源，按题目稳定顺序返回题目序列
func (r *MistakeRepository) ListQuestions(userID uint) ([]model.Question, error) {
	var qs []model.Question
	sub := r.db.Model(&model.Mistake{}).Select("question_id").Where("user_id = ?", userID)
	err := r.db.Where("id IN (?)", sub).Order("id").Find(&qs).Error
	return qs, err
}

// Delete 手动移除错题；目标不存在时不报错（幂等）
func (r *MistakeRepository) Delete(userID, questionID uint) error {
	return r.db.Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.Mistake{}).Error
}
