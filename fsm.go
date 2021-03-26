package main

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
