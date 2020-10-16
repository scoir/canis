package rabbitmq

import (
	"log"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Publisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

func NewPublisher(addr, queue string) (*Publisher, error) {
	var err error
	conn, err := amqp.Dial(addr)
	if err != nil {
		log.Fatalln("unable to dial AMQP", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create an AMQP channel")
	}

	_, err = ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to declare AMQP queue")
	}

	return &Publisher{
		conn:  conn,
		ch:    ch,
		queue: queue,
	}, nil
}

func (r *Publisher) Publish(body []byte, contentType string) error {
	err := r.ch.Publish(
		"",      // exchange
		r.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: contentType,
			Body:        body,
		})

	return errors.Wrap(err, "rabbitMQ publish failed")

}

func (r *Publisher) Close() error {
	return r.conn.Close()
}
