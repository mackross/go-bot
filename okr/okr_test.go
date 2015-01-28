package okr

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gorhill/cronexpr"
)

func TestBetween(t *testing.T) {
	weekdays := cronexpr.MustParse("0 9 * * 1-5")
	start := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2015, 3, 1, 0, 0, 0, 0, time.UTC)

	expected := []int64{1420102800, 1420189200, 1420448400, 1420534800, 1420621200, 1420707600, 1420794000, 1421053200, 1421139600, 1421226000, 1421312400, 1421398800, 1421658000, 1421744400, 1421830800, 1421917200, 1422003600, 1422262800, 1422349200, 1422435600, 1422522000, 1422608400, 1422867600, 1422954000, 1423040400, 1423126800, 1423213200, 1423472400, 1423558800, 1423645200, 1423731600, 1423818000, 1424077200, 1424163600, 1424250000, 1424336400, 1424422800, 1424682000, 1424768400, 1424854800, 1424941200, 1425027600}

	results := make([]int64, 0, len(expected))
	for _, t := range between(start, end, weekdays) {
		results = append(results, t.Unix())
	}

	equals(t, expected, results)
}

func TestGenerateQuestionsAfterTime(t *testing.T) {
	feb2nd2015 := time.Date(2015, 2, 2, 0, 0, 0, 0, time.UTC)
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"

	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, nil, TextAnswerType()}
	questions, err := spec.generateQuestions(feb2nd2015)
	ok(t, err)

	expected := []Question{Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 3, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}}
	equals(t, expected, questions)

}

func TestGenerateQuestionsBeforeStartTime(t *testing.T) {
	oct21st2014 := time.Date(2014, 10, 21, 0, 0, 0, 0, time.UTC)
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"

	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, nil, TextAnswerType()}
	questions, err := spec.generateQuestions(oct21st2014)
	ok(t, err)

	expected := []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 3, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}}
	equals(t, expected, questions)

	fmt.Printf("%#v\n%#v\n", questions, expected)
}

func TestRemoveAskedAt(t *testing.T) {
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"

	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: ptrTime(time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC)), AnsweredAt: nil}}, TextAnswerType()}

	expected := []Question{Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: ptrTime(time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC)), AnsweredAt: nil}}

	spec.removeUnaskedQuestions()
	equals(t, expected, spec.Questions)

}

func TestRange(t *testing.T) {
	a, b := RangeAnswerType(2, 5).parseRange()
	equals(t, float64(2), a)
	equals(t, float64(5), b)
}

func TestCheckAnswer(t *testing.T) {
	tests := []struct {
		answer interface{}
		qs     QuestionSpec
		err    error
	}{
		{
			answer: "blah blah",
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, TextAnswerType()},
			err:    nil,
		},
		{
			answer: "",
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, TextAnswerType()},
			err:    errors.New("answer must not be empty"),
		},
		{
			answer: 1,
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, TextAnswerType()},
			err:    errors.New("answer must be a string for text answers"),
		},
		{
			answer: float64(0),
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 5)},
			err:    nil,
		},
		{
			answer: "1",
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 5)},
			err:    errors.New("answer must be a number for range answers"),
		},
		{
			answer: float64(5.1),
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 5)},
			err:    errors.New("answer must be equal to or between 0 and 5"),
		},
		{
			answer: float64(3),
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 5)},
			err:    nil,
		},
		{
			answer: float64(5),
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 5)},
			err:    nil,
		},
		{
			answer: float64(-0.2),
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 4.2)},
			err:    errors.New("answer must be equal to or between 0 and 4.2"),
		},
		{
			answer: float64(0.2),
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, RangeAnswerType(0, 5)},
			err:    nil,
		},
		{
			answer: "true",
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, BoolAnswerType()},
			err:    errors.New("answer must be a boolean"),
		},
		{
			answer: true,
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, BoolAnswerType()},
			err:    nil,
		},
		{
			answer: false,
			qs:     QuestionSpec{"", "", time.Now(), time.Now(), nil, BoolAnswerType()},
			err:    nil,
		},
	}
	for i, test := range tests {
		err := test.qs.checkAnswer(test.answer)
		assert(t, reflect.DeepEqual(test.err, err), "test %v fail err (%v != %v)", i+1, err, test.err)
	}
}

func TestRemoveUnanswered(t *testing.T) {
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"

	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: ptrTime(time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC))}}, TextAnswerType()}

	expected := []Question{Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: ptrTime(time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC))}}

	spec.removeUnansweredQuestions()
	equals(t, expected, spec.Questions)

}

func TestLastAnsweredQuestion(t *testing.T) {
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"

	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: ptrTime(time.Date(2015, 2, 28, 9, 1, 0, 0, time.UTC))}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: ptrTime(time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC))}}, TextAnswerType()}

	expected := &Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: ptrTime(time.Date(2015, 2, 28, 9, 1, 0, 0, time.UTC))}

	q := spec.lastAnsweredQuestion()
	equals(t, q, expected)

	spec = QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}}, TextAnswerType()}
	assert(t, spec.lastAnsweredQuestion() == nil, "no question answered")
}

func TestUnaskedQuestionsBefore(t *testing.T) {
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"
	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: ptrTime(time.Date(2015, 2, 28, 9, 1, 0, 0, time.UTC)), AnsweredAt: nil}}, TextAnswerType()}

	expected := []*Question{&Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}}

	equals(t, expected, spec.unaskedQuestionsBefore(time.Date(2015, 10, 10, 9, 0, 0, 0, time.UTC)))
	equals(t, []*Question{}, spec.unaskedQuestionsBefore(time.Date(2000, 10, 10, 9, 0, 0, 0, time.UTC)))
	equals(t, expected, spec.unaskedQuestionsBefore(time.Date(2015, 2, 28, 8, 0, 0, 0, time.UTC)))

}
func TestAskedButUnansweredQuestions(t *testing.T) {
	jan1st2015 := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	april1st2015 := time.Date(2015, 4, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := "0 9 L * *"
	spec := QuestionSpec{"Did you have a fun month?", endOfMonth, jan1st2015, april1st2015, []Question{Question{Answer: "", AskAt: time.Date(2015, 1, 31, 9, 0, 0, 0, time.UTC), AskedAt: nil, AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: ptrTime(time.Date(2015, 2, 28, 9, 1, 0, 0, time.UTC)), AnsweredAt: nil}, Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: ptrTime(time.Date(2015, 2, 28, 9, 1, 0, 0, time.UTC)), AnsweredAt: ptrTime(time.Date(2015, 2, 28, 9, 2, 0, 0, time.UTC))}}, TextAnswerType()}

	expected := []*Question{&Question{Answer: "", AskAt: time.Date(2015, 2, 28, 9, 0, 0, 0, time.UTC), AskedAt: ptrTime(time.Date(2015, 2, 28, 9, 1, 0, 0, time.UTC)), AnsweredAt: nil}}

	equals(t, expected, spec.unansweredButAskedQuestions())

}

func ptrTime(t time.Time) *time.Time {
	return &t
}
