package service

import (
	"math/rand/v2"

	"software-exam/backend/internal/model"
)

// shuffleQuestions 以 Fisher–Yates 打乱题目并返回副本，不改动入参切片。
// intn 注入 [0,n) 随机源便于单测，传 nil 用 math/rand/v2 全局源
func shuffleQuestions(questions []model.Question, intn func(n int) int) []model.Question {
	if intn == nil {
		intn = rand.IntN
	}
	shuffled := append([]model.Question(nil), questions...)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}
