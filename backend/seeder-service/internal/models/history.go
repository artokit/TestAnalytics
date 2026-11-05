package models

type History struct {
	Subs     []Subscription
	Payments []Payment
}

func NewHistory() *History {
	return &History{
		Subs:     make([]Subscription, 0),
		Payments: make([]Payment, 0),
	}
}
