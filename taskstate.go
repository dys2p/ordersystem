package ordersystem

import (
	"fmt"
	"html/template"
)

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

func (s TaskState) Name() string {
	switch s {
	case Failed:
		return "Gescheitert"
	case Fetched:
		return "Abgeholt"
	case NotOrderedYet:
		return "Noch nicht bestellt"
	case Ordered:
		return "Bestellt"
	case Ready:
		return "Eingetroffen"
	case Reshipped:
		return "Weiterverschickt"
	case Unfetched:
		return "Nicht abgeholt"
	default:
		return string(s)
	}
}

func (s TaskState) Description() template.HTML {
	var desc string
	switch s {
	case Failed:
		desc = "(Die Ausführung dieser Bestellung ist gescheitert.)"
	case Fetched:
		desc = "(Die Ware wurde abgeholt.)"
	case NotOrderedYet:
		desc = "(Wir haben die Bestellung noch nicht ausgeführt.)"
	case Ordered:
		desc = "(Wir haben deinen Bestellauftrag ausgeführt. Du wirst benachrichtigt, wenn die Ware eintrifft.)"
	case Ready:
		desc = "(Die Ware ist bei uns eingetroffen. Du kannst sie abholen bzw. wir versenden sie bald an dich weiter.)"
	case Reshipped:
		desc = "(Wir haben deine Ware an dich weiterverschickt.)"
	case Unfetched:
		desc = "(Deine Ware wurde nicht abgeholt. Bitte melde dich bei uns.)"
	}
	return template.HTML(fmt.Sprintf("<strong>%s</strong> %s", s.Name(), desc))
}
