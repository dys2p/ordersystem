package main

type TaskState string

const (
	Failed        TaskState = "failed"
	Fetched       TaskState = "fetched"
	NotOrderedYet TaskState = "not-ordered-yet" // initial state
	Ordered       TaskState = "ordered"
	Ready         TaskState = "ready"
	Reshipped     TaskState = "reshipped"
	Unfetched     TaskState = "unfetched"
)

// for translation
func (s TaskState) NameKey() string {
	return "taskstate-" + string(s)
}

// for translation
func (s TaskState) DescriptionKey() string {
	return "taskstate-desc-" + string(s)
}
