package main

import (
	"fmt"
	"html/template"
)

type CollState string

const (
	Accepted    CollState = "accepted"
	Archived    CollState = "archived"
	Cancelled   CollState = "cancelled"
	Deleted     CollState = "deleted"
	Draft       CollState = "draft"
	Finalized   CollState = "finalized"
	NeedsRevise CollState = "needs-revise"
	Paid        CollState = "paid"
	Rejected    CollState = "rejected"
	Spam        CollState = "spam"
	Submitted   CollState = "submitted"
	Underpaid   CollState = "underpaid"
	// no "overpaid" state, just return the money when the goods are fetched or shipped
)

func (s CollState) Name() string {
	switch s {
	case Accepted:
		return "Angenommen"
	case Archived:
		return "Archiviert"
	case Cancelled:
		return "Abgebrochen"
	case Draft:
		return "Entwurf"
	case Finalized:
		return "Abgeschlossen"
	case NeedsRevise:
		return "Erfordert Überarbeitung"
	case Paid:
		return "Bezahlt, wird ausgeführt"
	case Rejected:
		return "Abgelehnt"
	case Spam:
		return "Spam"
	case Submitted:
		return "Eingereicht"
	case Underpaid:
		return "Unterbezahlt"
	default:
		return string(s)
	}
}

func (s CollState) Description() template.HTML {
	var desc string
	switch s {
	case Accepted:
		desc = "(Dein Bestellauftrag wurde angenommen. Bitte lasse uns das Geld zukommen. Dann führen wir den Bestellauftrag aus.)"
	case Archived:
		desc = "(Kontakt- und Lieferinformationen wurden gelöscht.)"
	case Cancelled:
		desc = "(Dein Bestellauftrag wurde abgebrochen.)"
	case Draft:
		desc = "(Du kannst deinen Bestellauftrag bearbeiten. Wenn du fertig bist, dann reiche ihn ein.)"
	case Finalized:
		desc = "(Die Ware wurde vollständig abgeholt oder verschickt.)"
	case NeedsRevise:
		desc = "(Wir haben eine Rückfrage zu deinem Bestellauftrag.)"
	case Paid:
		desc = "(Wir haben dein Geld erhalten und sind dabei, deine Bestellaufträge auszuführen.)"
	case Rejected:
		desc = "(Wir können deinen Bestellauftrag leider nicht ausführen.)"
	case Submitted:
		desc = "(Du hast deinen Bestellauftrag abgeschickt. Er wird jetzt von uns geprüft.)"
	case Underpaid:
		desc = "(Bitte lasse uns den Differenzbetrag zukommen oder teile uns mit, ob wir einen Artikel entfernen sollen.)"
	}
	return template.HTML(fmt.Sprintf("<strong>%s</strong> %s", s.Name(), desc))
}
