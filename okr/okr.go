package okr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

type Repo interface {
}

var now = func() time.Time {
	return time.Now()
}

type OKR struct {
	Title         string
	UserID        string
	ID            string
	QuestionSpecs []QuestionSpec
}

type answerType string

func RangeAnswerType(lower float64, upper float64) answerType {
	return answerType(fmt.Sprintf("range[%v:%v]", lower, upper))
}

func TextAnswerType() answerType {
	return answerType("text")
}
func BoolAnswerType() answerType {
	return answerType("bool")
}

func (a answerType) isTextAnswer() bool {
	return a == TextAnswerType()
}

func (a answerType) isRangeAnswer() bool {
	return strings.HasPrefix(string(a), "range")
}

func (a answerType) isBoolAnswer() bool {
	return a == BoolAnswerType()
}

func (r answerType) parseRange() (float64, float64) {
	str := strings.Split(strings.TrimRight(strings.TrimLeft(string(r), "range["), "]"), ":")
	lower, err := strconv.ParseFloat(str[0], 64)
	check(err)
	upper, err := strconv.ParseFloat(str[1], 64)
	check(err)
	return lower, upper
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type QuestionSpec struct {
	Question   string
	Schedule   string
	Starts     time.Time
	Ends       time.Time
	Questions  []Question
	AnswerType answerType
}

type Question struct {
	Answer     interface{}
	AskAt      time.Time
	AskedAt    *time.Time
	AnsweredAt *time.Time
}

func (qs QuestionSpec) checkAnswer(a interface{}) error {
	if qs.AnswerType.isTextAnswer() {
		if str, ok := a.(string); ok {
			if len(str) == 0 {
				return errors.New("answer must not be empty")
			}
			return nil
		}
		return errors.New("answer must be a string for text answers")
	} else if qs.AnswerType.isRangeAnswer() {
		if f, ok := a.(float64); ok {
			lower, upper := qs.AnswerType.parseRange()
			if f < lower || f > upper {
				return errors.New(fmt.Sprintf("answer must be equal to or between %v and %v", lower, upper))
			}
			return nil
		}
		return errors.New("answer must be a number for range answers")
	} else if qs.AnswerType.isBoolAnswer() {
		if _, ok := a.(bool); ok {
			return nil
		}
		return errors.New("answer must be a boolean")
	}
	return nil
}

func (s *QuestionSpec) removeUnaskedQuestions() {
	qs := make([]Question, 0)
	for _, q := range s.Questions {
		if q.AskedAt != nil {
			qs = append(qs, q)
		}
	}
	s.Questions = qs
}

func (s *QuestionSpec) removeUnansweredQuestions() {
	qs := make([]Question, 0)
	for _, q := range s.Questions {
		if q.AnsweredAt != nil {
			qs = append(qs, q)
		}
	}
	s.Questions = qs
}

func (s *QuestionSpec) lastAnsweredQuestion() *Question {
	idx := -1
	for i, q := range s.Questions {
		if q.AnsweredAt != nil {
			neverSetOrLater := idx == -1 || q.AnsweredAt.After(*(s.Questions[i].AnsweredAt))
			if neverSetOrLater {
				idx = i
			}
		}
	}
	if idx == -1 {
		return nil
	}
	return &s.Questions[idx]
}

func (s *QuestionSpec) unaskedQuestionsBefore(t time.Time) []*Question {
	qs := make([]*Question, 0)
	for i := 0; i < len(s.Questions); i++ {
		q := s.Questions[i]
		if q.AskedAt == nil {
			if q.AskAt.Before(t) {
				qs = append(qs, &q)
			}
		}
	}
	return qs
}

func (s *QuestionSpec) unansweredButAskedQuestions() []*Question {
	qs := make([]*Question, 0)
	for i := 0; i < len(s.Questions); i++ {
		q := s.Questions[i]
		if q.AskedAt != nil && q.AnsweredAt == nil {
			qs = append(qs, &q)
		}
	}
	return qs
}

func (s *QuestionSpec) generateQuestions(startTime time.Time) ([]Question, error) {
	expr, err := cronexpr.Parse(s.Schedule)
	if err != nil {
		return nil, err
	}
	times := between(later(startTime, s.Starts), s.Ends, expr)
	fmt.Println("times:", times)
	questions := make([]Question, 0, len(times))

	for _, t := range times {
		q := Question{"", t, nil, nil}
		questions = append(questions, q)
	}
	return questions, nil
}

func between(s time.Time, e time.Time, expr *cronexpr.Expression) []time.Time {
	times := make([]time.Time, 0)
	t := s
	for {
		t = expr.Next(t)
		if t.After(e) {
			return times
		}
		times = append(times, t)
		t = t.Add(time.Second)
	}
}

func later(t1 time.Time, t2 time.Time) time.Time {
	if t1.After(t2) {
		return t1
	}
	return t2
}
