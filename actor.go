package ordersystem

type Actor string

const (
	Bot    Actor = "bot"
	Client Actor = "client"
	Store  Actor = "store"
)

// for templates, so they don't have to use string literals
func (a Actor) IsClient() bool {
	return a == Client
}

// for templates, so they don't have to use string literals
func (a Actor) IsStore() bool {
	return a == Store
}

func (a Actor) Name() string {
	switch a {
	case Bot:
		return "Bot"
	case Client:
		return "Client"
	case Store:
		return "Store"
	default:
		return string(a)
	}
}
