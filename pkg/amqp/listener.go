package amqp

import (
	"github.com/streadway/amqp"
)

//go:generate mockery -name=Listener
type Listener interface {
	Listen() (<-chan amqp.Delivery, error)
}
