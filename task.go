package main

import (
	"math"
)

const storeFeeBase = 1290
const storeFeeShare = 0.02

type Task struct {
	// stored as SQL row, but partly unmarshaled from user input, hence the JSON tags
	ID    string    `json:"id"`
	State TaskState `json:"-"`
	TaskData
}

func (task *Task) Writeable(actor Actor) bool {
	switch actor {
	case Bot:
		return false
	case Client:
		s := task.State
		return s == Failed || s == NotOrderedYet
	case Store:
		s := task.State
		return s == Failed || s == NotOrderedYet || s == Ordered
	default:
		return false
	}
}

func (task *Task) Fee() int {
	return int(math.Round(storeFeeShare*float64(task.Sum()))) + storeFeeBase
}

// Sum is the sum of articles, shipping fee and additional costs. No store fee included.
func (task *Task) Sum() int {
	var sum = 0
	for _, a := range task.Articles {
		if a.Quantity > 0 && a.Price > 0 {
			sum += a.Quantity * a.Price
		}
	}
	sum += task.ShippingFee
	for _, addCost := range task.AddCosts {
		sum += addCost.Price
	}
	return sum
}

// TaskData is a separate struct so we can marshal it easily and store it in the SQL database.
type TaskData struct {
	AddCosts    []AddCost `json:"add-costs"`
	Articles    []Article `json:"articles"`
	Merchant    string    `json:"merchant"`
	ShippingFee int       `json:"shipping-fee"`
}

type Article struct {
	Link       string `json:"link"`
	Price      int    `json:"price"`
	Properties string `json:"properties"`
	Quantity   int    `json:"quantity"`
}

type AddCost struct {
	Name  string `json:"name"`
	Price int    `json:"price"` // positive (expenses) or negative (discount)
}
