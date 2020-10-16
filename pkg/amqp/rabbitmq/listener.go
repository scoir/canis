package rabbitmq

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Listener struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

func NewListener(addr, queue string) (*Listener, error) {
	l := &Listener{queue: queue}

	var err error
	l.conn, err = amqp.Dial(addr)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to RabbitMQ at %s", addr)
	}

	l.ch, err = l.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get channel")
	}

	_, err = l.ch.QueueDeclare(
		queue,
		false,
		false,
		false,
		false,
		nil,
	)

	return l, err
}

func (r *Listener) Listen() (<-chan amqp.Delivery, error) {
	msgs, err := r.ch.Consume(
		r.queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to consume")
	}

	return msgs, nil
}

func (r *Listener) Close() error {
	return r.conn.Close()
}
