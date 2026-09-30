package service

import (
	"reflect"
	"slices"
	"testing"

	"software-exam/backend/internal/model"
)

func testQuestions(n int) []model.Question {
	questions := make([]model.Question, n)
	for i := range questions {
		questions[i] = model.Question{ID: uint(i + 1), No: i + 1}
	}
	return questions
}

func questionIDs(questions []model.Question) []uint {
	ids := make([]uint, len(questions))
	for i, q := range questions {
		ids[i] = q.ID
	}
	return ids
}

// 注入恒返 0 的随机源后，Fisher–Yates 退化为确定排列，可精确断言结果
func TestShuffleQuestionsDeterministicStub(t *testing.T) {
	got := shuffleQuestions(testQuestions(4), func(int) int { return 0 })
	want := []uint{2, 3, 4, 1}
	if !reflect.DeepEqual(questionIDs(got), want) {
		t.Fatalf("恒 0 随机源下 ID 序列 = %v, want %v", questionIDs(got), want)
	}
}

func TestShuffleQuestionsPreservesElements(t *testing.T) {
	questions := testQuestions(20)
	originalOrder := questionIDs(questions)
	got := shuffleQuestions(questions, nil)

	sortedGot := slices.Sorted(slices.Values(questionIDs(got)))
	sortedWant := slices.Sorted(slices.Values(originalOrder))
	if !reflect.DeepEqual(sortedGot, sortedWant) {
		t.Fatalf("打乱后题目集合改变: got %v, want %v", sortedGot, sortedWant)
	}

	// 返回副本，入参切片保持原序
	if !slices.Equal(questionIDs(questions), originalOrder) {
		t.Fatalf("入参切片被改动: %v", questionIDs(questions))
	}
}

func TestShuffleQuestionsActuallyShuffles(t *testing.T) {
	// 默认随机源下 20 题全等原序的概率可忽略；多次取样至少出现一次乱序
	questions := testQuestions(20)
	for range 50 {
		got := shuffleQuestions(questions, nil)
		if !slices.Equal(questionIDs(got), questionIDs(questions)) {
			return
		}
	}
	t.Fatal("默认随机源下 50 次取样均为原序，疑似未打乱")
}

func TestShuffleQuestionsEdgeCases(t *testing.T) {
	for _, n := range []int{0, 1} {
		got := shuffleQuestions(testQuestions(n), nil)
		if len(got) != n {
			t.Fatalf("n=%d 打乱后长度 = %d", n, len(got))
		}
	}
}
