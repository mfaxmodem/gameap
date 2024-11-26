package entity

type Question struct {
	ID              uint
	Text            string
	PossibleAnswers []string
	CorrectAnswerID uint
	Difficulty      QuestionDifficulty
	CategoryID      uint
}

type PossibleAnswers struct {
	ID     uint
	Text   string
	Choice PossibleAnswersChoice
}

type PossibleAnswersChoice int8

func (p PossibleAnswersChoice) IsValid() bool {
	if p >= PossibleAnswersA && p <= PossibleAnswersD {
		return true
	}
	return false
}

const (
	PossibleAnswersA PossibleAnswersChoice = iota + 1
	PossibleAnswersB
	PossibleAnswersC
	PossibleAnswersD
)

type QuestionDifficulty int8

const (
	QuestionDifficultyEasy QuestionDifficulty = iota + 1
	QuestionDifficultyMedium
	QuestionDifficultyHard
)

func (q QuestionDifficulty) IsValid() bool {
	if q >= QuestionDifficultyEasy && q <= QuestionDifficultyHard {
		return true
	}
	return false
}
