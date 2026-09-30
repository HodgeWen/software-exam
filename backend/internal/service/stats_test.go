package service

import (
	"testing"

	"software-exam/backend/internal/repository"
)

func strPtr(s string) *string { return &s }
func uintPtr(v uint) *uint    { return &v }

// TestAggregateStatsMultiSubjectChapter 统计计算：多科目/多章节数据折成总量、正确率与分布，
// 真题试卷题（无章节）计入科目不计入章节分布
func TestAggregateStatsMultiSubjectChapter(t *testing.T) {
	chA1, chA2, chB1 := uintPtr(11), uintPtr(12), uintPtr(21)
	dims := []repository.AnswerDim{
		// 科目 A 章节一：3 答 2 对
		{SubjectID: 1, SubjectName: "科目A", ChapterID: chA1, ChapterName: strPtr("章节一"), Correct: true},
		{SubjectID: 1, SubjectName: "科目A", ChapterID: chA1, ChapterName: strPtr("章节一"), Correct: false},
		{SubjectID: 1, SubjectName: "科目A", ChapterID: chA1, ChapterName: strPtr("章节一"), Correct: true},
		// 科目 A 章节二：2 答 0 对
		{SubjectID: 1, SubjectName: "科目A", ChapterID: chA2, ChapterName: strPtr("章节二"), Correct: false},
		{SubjectID: 1, SubjectName: "科目A", ChapterID: chA2, ChapterName: strPtr("章节二"), Correct: false},
		// 科目 A 真题试卷题：2 答 1 对，只进科目不进章节
		{SubjectID: 1, SubjectName: "科目A", Correct: false},
		{SubjectID: 1, SubjectName: "科目A", Correct: true},
		// 科目 B 章节三：1 答 1 对
		{SubjectID: 2, SubjectName: "科目B", ChapterID: chB1, ChapterName: strPtr("章节三"), Correct: true},
	}

	s := aggregateStats(dims)

	if s.Total != 8 || s.Correct != 4 || s.Accuracy != 0.5 {
		t.Fatalf("总量/正确率错误: total=%d correct=%d accuracy=%v", s.Total, s.Correct, s.Accuracy)
	}
	if len(s.Subjects) != 2 {
		t.Fatalf("科目数 = %d, want 2: %+v", len(s.Subjects), s.Subjects)
	}
	a, b := s.Subjects[0], s.Subjects[1]
	if a.SubjectID != 1 || a.SubjectName != "科目A" || a.Total != 7 || a.Correct != 3 || a.Accuracy != 0.4286 {
		t.Fatalf("科目 A 统计错误: %+v", a)
	}
	if len(a.Chapters) != 2 {
		t.Fatalf("科目 A 章节数 = %d, want 2: %+v", len(a.Chapters), a.Chapters)
	}
	if a.Chapters[0].ChapterID != 11 || a.Chapters[0].ChapterName != "章节一" ||
		a.Chapters[0].Total != 3 || a.Chapters[0].Correct != 2 || a.Chapters[0].Accuracy != 0.6667 {
		t.Fatalf("章节一统计错误: %+v", a.Chapters[0])
	}
	if a.Chapters[1].ChapterID != 12 || a.Chapters[1].Total != 2 ||
		a.Chapters[1].Correct != 0 || a.Chapters[1].Accuracy != 0 {
		t.Fatalf("章节二统计错误: %+v", a.Chapters[1])
	}
	if b.SubjectID != 2 || b.Total != 1 || b.Correct != 1 || b.Accuracy != 1 {
		t.Fatalf("科目 B 统计错误: %+v", b)
	}
	if len(b.Chapters) != 1 || b.Chapters[0].ChapterID != 21 || b.Chapters[0].Total != 1 {
		t.Fatalf("科目 B 章节分布错误: %+v", b.Chapters)
	}
}

// TestAggregateStatsEmpty 无作答时全 0，subjects 为空数组而非 null
func TestAggregateStatsEmpty(t *testing.T) {
	s := aggregateStats(nil)
	if s.Total != 0 || s.Correct != 0 || s.Accuracy != 0 {
		t.Fatalf("空数据应全 0: %+v", s)
	}
	if s.Subjects == nil || len(s.Subjects) != 0 {
		t.Fatalf("空数据 subjects 应为空数组: %+v", s.Subjects)
	}
}
