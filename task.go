package ordersystem

import (
	"math"
)

const storeFeeBase = 1000
const storeFeeShare = 0.05

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
		return s == Failed || s == NotOrderedYet || s == Ordered || s == Ready || s == Fetched || s == Reshipped
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

func (task *Task) TotalSum() int {
	return task.Sum() + task.Fee()
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
	Price      int    `json:"price"` // item price
	Properties string `json:"properties"`
	Quantity   int    `json:"quantity"`
}

func (article Article) Sum() int {
	return article.Quantity * article.Price
}

type AddCost struct {
	Name  string `json:"name"`
	Price int    `json:"price"` // positive (expenses) or negative (discount)
}
