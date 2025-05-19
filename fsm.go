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
	Transition{State(Accepted), Bot, "confirm-payment", State(Active)},
	Transition{State(Accepted), Bot, "delete", State(Deleted)},
	Transition{State(Accepted), Client, "cancel", State(Cancelled)},
	Transition{State(Accepted), Client, "pay", State(Accepted)}, // becomes Paid if payment arrives
	Transition{State(Accepted), Store, "confirm-payment", State(Active)},
	Transition{State(Accepted), Store, "delete", State(Deleted)},
	Transition{State(Accepted), Store, "edit", State(Accepted)},
	Transition{State(Accepted), Store, "activate", State(Active)},
	Transition{State(Accepted), Store, "return", State(NeedsRevise)},
	Transition{State(Draft), Bot, "delete", State(Deleted)},
	Transition{State(Draft), Client, "delete", State(Deleted)},
	Transition{State(Draft), Client, "edit", State(Draft)},
	Transition{State(Draft), Client, "submit", State(Submitted)},
	Transition{State(Draft), Store, "submit", State(Submitted)},
	Transition{State(Submitted), Client, "message", State(Submitted)},
	Transition{State(Submitted), Store, "message", State(Submitted)},
	Transition{State(Finalized), Store, "message", State(Finalized)}, // "Hi, we just shipped your order."
	Transition{State(Finalized), Bot, "archive", State(Archived)},
	Transition{State(NeedsRevise), Client, "cancel", State(Cancelled)},
	Transition{State(NeedsRevise), Client, "edit", State(NeedsRevise)},
	Transition{State(NeedsRevise), Client, "submit", State(Submitted)},
	Transition{State(Active), Bot, "confirm-payment", State(Active)},
	Transition{State(Active), Bot, "finalize", State(Finalized)},
	Transition{State(Active), Store, "confirm-payment", State(Accepted)}, // all tasks failed
	Transition{State(Active), Store, "confirm-payment", State(Active)},   // refund overpaid amount
	Transition{State(Active), Store, "confirm-pickup", State(Active)},
	Transition{State(Active), Store, "confirm-reshipped", State(Active)},
	Transition{State(Active), Store, "edit", State(Active)}, // price or availability changed after payment
	Transition{State(Active), Store, "message", State(Active)},
	Transition{State(Spam), Bot, "delete", State(Deleted)},
	Transition{State(Submitted), Client, "cancel", State(Cancelled)},
	Transition{State(Submitted), Store, "accept", State(Accepted)},
	Transition{State(Submitted), Store, "edit", State(Submitted)},
	Transition{State(Submitted), Store, "mark-spam", State(Spam)},
	Transition{State(Submitted), Store, "reject", State(Rejected)},
	Transition{State(Submitted), Store, "return", State(NeedsRevise)},
	Transition{State(Active), Bot, "confirm-payment", State(Active)},
	Transition{State(Active), Client, "message", State(Active)},
	Transition{State(Active), Client, "pay", State(Active)},              // becomes Paid if payment arrives
	Transition{State(Active), Store, "confirm-payment", State(Accepted)}, // store refunds whole amount
	Transition{State(Active), Store, "confirm-payment", State(Active)},   // client pays missing amount or a part of it
	Transition{State(Active), Store, "edit", State(Active)},              // store modifies the collection
	Transition{State(Active), Store, "message", State(Active)},
	Transition{State(Finalized), Store, "activate", State(Active)},
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
