package main

type CollState string

const (
	Accepted    CollState = "accepted"
	Archived    CollState = "archived"
	Cancelled   CollState = "cancelled"
	Deleted     CollState = "deleted" // used for the transition only
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

// for translation
func (s CollState) NameKey() string {
	return "collstate-" + string(s)
}

// for translation
func (s CollState) DescriptionKey() string {
	return "collstate-desc-" + string(s)
}
