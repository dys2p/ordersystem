package ordersystem

type State string

type Transition struct {
	from   State
	actor  Actor
	action string // not always necessary
	to     State
}

type FSM []Transition // easier than a map for a small number of transactions

func (fsm *FSM) Can(me Actor, from State, to State) bool {
	for _, t := range *fsm {
		if t.actor == me && t.from == from && t.to == to {
			return true
		}
	}
	return false
}

func (fsm *FSM) CanAction(me Actor, from State, action string) bool {
	for _, t := range *fsm {
		if t.actor == me && t.from == from && t.action == action {
			return true
		}
	}
	return false
}

func (fsm *FSM) From(me Actor) []State {
	var from = []State{}
	for _, t := range *fsm {
		if t.actor == me {
			from = append(from, t.from)
		}
	}
	return from
}

var CollFSM = &FSM{
	Transition{State(Accepted), Bot, "confirm-payment", State(Paid)},
	Transition{State(Accepted), Bot, "confirm-payment", State(Underpaid)},
	Transition{State(Accepted), Bot, "delete", State(Deleted)},
	Transition{State(Accepted), Client, "cancel", State(Cancelled)},
	Transition{State(Accepted), Client, "pay", State(Accepted)},             // becomes Paid if payment arrives
	Transition{State(Accepted), Store, "confirm-payment", State(Paid)},      // client pays enough
	Transition{State(Accepted), Store, "confirm-payment", State(Underpaid)}, // client pays, but not enough
	Transition{State(Accepted), Store, "delete", State(Deleted)},
	Transition{State(Accepted), Store, "edit", State(Accepted)},
	Transition{State(Accepted), Store, "return", State(NeedsRevise)},
	Transition{State(Draft), Bot, "delete", State(Deleted)},
	Transition{State(Draft), Client, "delete", State(Deleted)},
	Transition{State(Draft), Client, "edit", State(Draft)},
	Transition{State(Draft), Client, "submit", State(Submitted)},
	Transition{State(Draft), Store, "submit", State(Submitted)},
	Transition{State(Finalized), Store, "message", State(Finalized)}, // "Hi, we just shipped your order."
	Transition{State(Finalized), Bot, "archive", State(Archived)},
	Transition{State(NeedsRevise), Client, "cancel", State(Cancelled)},
	Transition{State(NeedsRevise), Client, "edit", State(NeedsRevise)},
	Transition{State(NeedsRevise), Client, "submit", State(Submitted)},
	Transition{State(Paid), Bot, "confirm-payment", State(Paid)},
	Transition{State(Paid), Bot, "finalize", State(Finalized)},
	Transition{State(Paid), Store, "confirm-payment", State(Accepted)}, // all tasks failed
	Transition{State(Paid), Store, "confirm-payment", State(Paid)},     // refund overpaid amount
	Transition{State(Paid), Store, "confirm-pickup", State(Paid)},
	Transition{State(Paid), Store, "confirm-reshipped", State(Paid)},
	Transition{State(Paid), Store, "edit", State(Paid)}, // price or availability changed after payment
	Transition{State(Paid), Store, "message", State(Paid)},
	Transition{State(Paid), Store, "price-rised", State(Underpaid)},
	Transition{State(Spam), Bot, "delete", State(Deleted)},
	Transition{State(Submitted), Client, "cancel", State(Cancelled)},
	Transition{State(Submitted), Store, "accept", State(Accepted)},
	Transition{State(Submitted), Store, "edit", State(Submitted)},
	Transition{State(Submitted), Store, "mark-spam", State(Spam)},
	Transition{State(Submitted), Store, "reject", State(Rejected)},
	Transition{State(Submitted), Store, "return", State(NeedsRevise)},
	Transition{State(Underpaid), Bot, "confirm-payment", State(Paid)},
	Transition{State(Underpaid), Bot, "confirm-payment", State(Underpaid)},
	Transition{State(Underpaid), Client, "message", State(Underpaid)},
	Transition{State(Underpaid), Client, "pay", State(Underpaid)},            // becomes Paid if payment arrives
	Transition{State(Underpaid), Store, "confirm-payment", State(Accepted)},  // store refunds whole amount
	Transition{State(Underpaid), Store, "confirm-payment", State(Paid)},      // client pays missing amount
	Transition{State(Underpaid), Store, "confirm-payment", State(Underpaid)}, // client pays a part of the missing amount
	Transition{State(Underpaid), Store, "edit", State(Paid)},                 // store modifies the collection, the sum drops, paid sum is now enough
	Transition{State(Underpaid), Store, "edit", State(Underpaid)},            // store modifies the collection, but it is still underpaid
	Transition{State(Underpaid), Store, "message", State(Underpaid)},
}

var TaskFSM = &FSM{
	Transition{State(NotOrderedYet), Store, "confirm-ordered", State(Ordered)},
	Transition{State(NotOrderedYet), Store, "mark-failed", State(Failed)},
	Transition{State(Ordered), Store, "confirm-arrived", State(Ready)},
	Transition{State(Ordered), Store, "mark-failed", State(Failed)},
	Transition{State(Ready), Bot, "pickup-expired", State(Unfetched)}, // TODO
	Transition{State(Ready), Store, "confirm-pickup", State(Fetched)},
	Transition{State(Ready), Store, "confirm-reshipped", State(Reshipped)},
}
