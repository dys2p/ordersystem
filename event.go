package ordersystem

import (
	"html/template"

	"gitlab.com/golang-commonmark/markdown"
)

type Event struct {
	NewState CollState
	Date     Date
	Paid     int    // euro cents, adds to old values, positive amounts were paid by the client, negative amounts were paid by the store
	Text     string // CommonMark markdown
}

func (e *Event) TextHTML() template.HTML {
	return template.HTML(markdown.New().RenderToString([]byte(e.Text)))
}
