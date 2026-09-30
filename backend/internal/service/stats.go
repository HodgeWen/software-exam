package service

import (
	"fmt"
	"math"

	"software-exam/backend/internal/repository"
)

// ChapterStats 科目下按章节的答题分布
type ChapterStats struct {
	ChapterID   uint    `json:"chapter_id"`
	ChapterName string  `json:"chapter_name"`
	Total       int     `json:"total"`
	Correct     int     `json:"correct"`
	Accuracy    float64 `json:"accuracy"`
}

// SubjectStats 按科目的答题数/正确数/正确率与章节分布
type SubjectStats struct {
	SubjectID   uint           `json:"subject_id"`
	SubjectName string         `json:"subject_name"`
	Total       int            `json:"total"`
	Correct     int            `json:"correct"`
	Accuracy    float64        `json:"accuracy"`
	Chapters    []ChapterStats `json:"chapters"`
}

// StatsSummary 总答题量、总正确率与按科目（含章节）的分布
type StatsSummary struct {
	Total    int            `json:"total"`
	Correct  int            `json:"correct"`
	Accuracy float64        `json:"accuracy"`
	Subjects []SubjectStats `json:"subjects"`
}

// StatsService 刷题统计：基于 answer_record 聚合，按用户隔离
type StatsService struct {
	records *repository.AnswerRecordRepository
}

func NewStatsService(records *repository.AnswerRecordRepository) *StatsService {
	return &StatsService{records: records}
}

// Summary 汇总当前用户的答题量、正确率与科目/章节分布
func (s *StatsService) Summary(userID uint) (*StatsSummary, error) {
	dims, err := s.records.ListDims(userID)
	if err != nil {
		return nil, fmt.Errorf("查询答题维度: %w", err)
	}
	return aggregateStats(dims), nil
}

// aggregateStats 维度行折成统计：单趟累计总量与各科目/章节桶，再统一算正确率。
// 正确率四舍五入到 4 位小数，保证响应值稳定可断言；展示格式交给前端
func aggregateStats(dims []repository.AnswerDim) *StatsSummary {
	summary := &StatsSummary{Subjects: []SubjectStats{}}
	subIdx := make(map[uint]int)
	chapIdx := make(map[uint]map[uint]int)
	for _, d := range dims {
		summary.Total++
		if d.Correct {
			summary.Correct++
		}
		si, ok := subIdx[d.SubjectID]
		if !ok {
			summary.Subjects = append(summary.Subjects, SubjectStats{
				SubjectID: d.SubjectID, SubjectName: d.SubjectName, Chapters: []ChapterStats{},
			})
			si = len(summary.Subjects) - 1
			subIdx[d.SubjectID] = si
			chapIdx[d.SubjectID] = make(map[uint]int)
		}
		sub := &summary.Subjects[si]
		sub.Total++
		if d.Correct {
			sub.Correct++
		}
		// 真题试卷题不属任何章节：计入科目与总量，不进章节分布
		if d.ChapterID == nil {
			continue
		}
		ci, ok := chapIdx[d.SubjectID][*d.ChapterID]
		if !ok {
			ch := ChapterStats{ChapterID: *d.ChapterID}
			if d.ChapterName != nil {
				ch.ChapterName = *d.ChapterName
			}
			sub.Chapters = append(sub.Chapters, ch)
			ci = len(sub.Chapters) - 1
			chapIdx[d.SubjectID][*d.ChapterID] = ci
		}
		ch := &sub.Chapters[ci]
		ch.Total++
		if d.Correct {
			ch.Correct++
		}
	}
	summary.Accuracy = accuracyRate(summary.Total, summary.Correct)
	for i := range summary.Subjects {
		s := &summary.Subjects[i]
		s.Accuracy = accuracyRate(s.Total, s.Correct)
		for j := range s.Chapters {
			c := &s.Chapters[j]
			c.Accuracy = accuracyRate(c.Total, c.Correct)
		}
	}
	return summary
}

// accuracyRate 正确/总量的比值；无作答时为 0，避免除零
func accuracyRate(total, correct int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(correct)/float64(total)*1e4) / 1e4
}
