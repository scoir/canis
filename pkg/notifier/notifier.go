package notifier

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"github.com/scoir/canis/pkg/datastore"
)

type Server struct {
	store datastore.Store
	conn  *amqp.Connection
	ch    *amqp.Channel
}

type provider interface {
	GetDatastore() datastore.Store
	GetAMQPAddress() string
}

func New(prov provider) (*Server, error) {
	srv := &Server{
		store: prov.GetDatastore(),
	}

	var err error
	amqpaddr := prov.GetAMQPAddress()
	srv.conn, err = amqp.Dial(amqpaddr)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to RabbitMQ at %s", amqpaddr)
	}

	srv.ch, err = srv.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get channel")
	}

	_, err = srv.ch.QueueDeclare(
		QueueName,
		false,
		false,
		false,
		false,
		nil,
	)

	return srv, nil
}

func (r *Server) Start() error {
	return r.listenAndServe()
}

func (r *Server) listenAndServe() error {
	msgs, err := r.ch.Consume(
		QueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "unable to consume")
	}

	for d := range msgs {
		note := &Notification{}
		err := json.Unmarshal(d.Body, note)
		if err != nil {
			log.Println("bad notification message", err)
			continue
		}

		hooks, err := r.store.ListWebhooks(note.Topic)
		if err != nil {
			log.Printf("no webhooks for topic %s: (%v)", note.Topic, err)
		}

		event := &EventMessage{
			Event:     note.Event,
			Timestamp: time.Now().Unix(),
			EventData: note.EventData,
		}
		data, _ := json.Marshal(event)
		for _, hook := range hooks {
			resp, err := http.Post(hook.URL, "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("unable to post event to hook %s: (%v)", hook.URL, err)
				continue
			}

			if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
				b, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				log.Printf("error response from hook.  code: (%d): %s\n", resp.StatusCode, string(b))
			}
		}

	}

	return nil
}
