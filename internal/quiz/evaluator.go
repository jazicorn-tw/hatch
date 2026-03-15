package quiz

// Evaluator checks answers against questions using deterministic index comparison.
type Evaluator struct{}

// NewEvaluator returns an Evaluator.
func NewEvaluator() *Evaluator { return &Evaluator{} }

// Check returns true if answer (0-3) matches the question's CorrectIndex.
func (e *Evaluator) Check(q Question, answer int) bool {
	return answer == q.CorrectIndex
}
