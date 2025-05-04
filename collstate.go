package ordersystem

import (
	"fmt"
	"html/template"
)

type CollState string

const (
	Accepted    CollState = "accepted"
	Active      CollState = "active"
	Archived    CollState = "archived"
	Cancelled   CollState = "cancelled"
	Deleted     CollState = "deleted"
	Draft       CollState = "draft"
	Finalized   CollState = "finalized"
	NeedsRevise CollState = "needs-revise"
	Rejected    CollState = "rejected"
	Spam        CollState = "spam"
	Submitted   CollState = "submitted"
)

func (s CollState) Name() string {
	switch s {
	case Accepted:
		return "Angenommen"
	case Active:
		return "Wird ausgeführt"
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
	case Rejected:
		return "Abgelehnt"
	case Spam:
		return "Spam"
	case Submitted:
		return "Eingereicht"
	default:
		return string(s)
	}
}

func (s CollState) Description() template.HTML {
	var desc string
	switch s {
	case Accepted:
		desc = "(Dein Bestellauftrag wurde angenommen. Bitte lasse uns das Geld zukommen. Dann führen wir den Bestellauftrag aus.)"
	case Active:
		desc = "(Wir sind dabei, deine Bestellaufträge auszuführen.)"
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
	case Rejected:
		desc = "(Wir können deinen Bestellauftrag leider nicht ausführen.)"
	case Submitted:
		desc = "(Du hast deinen Bestellauftrag abgeschickt. Er wird jetzt von uns geprüft.)"
	}
	return template.HTML(fmt.Sprintf("<strong>%s</strong> %s", s.Name(), desc))
}
