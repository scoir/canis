package notifier

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/amqp"
	"github.com/scoir/canis/pkg/datastore"
)

type Server struct {
	store    datastore.Store
	listener amqp.Listener
	errors   chan error
}

type provider interface {
	GetDatastore() datastore.Store
	GetAMQPListener(queue string) amqp.Listener
}

func New(prov provider) (*Server, error) {
	srv := &Server{
		store:    prov.GetDatastore(),
		listener: prov.GetAMQPListener(QueueName),
	}

	return srv, nil
}

func (r *Server) Start() error {
	return r.listenAndServe()
}

func (r *Server) listenAndServe() error {
	msgs, err := r.listener.Listen()
	if err != nil {
		return errors.Wrap(err, "unable to consume")
	}

	for d := range msgs {
		note := &Notification{}
		err := json.Unmarshal(d.Body, note)
		if err != nil {
			r.Error(errors.Wrap(err, "bad notification message"))
			continue
		}

		hooks, err := r.store.ListWebhooks(note.Topic)
		if err != nil {
			r.Error(errors.Wrapf(err, "no webhooks for topic %s", note.Topic))
			continue
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
				r.Error(errors.Wrapf(err, "unable to post event to hook %s", hook.URL))
				continue
			}

			if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
				b, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				r.Error(errors.Errorf("error response from hook. code: (%d): %s\n", resp.StatusCode, string(b)))
			}
		}

	}

	return errors.New("notification messages closed")
}

func (r *Server) Error(err error) {
	if r.errors == nil {
		log.Println(err.Error())
	}

	r.errors <- err
}

func (r *Server) Errors() (chan error, error) {
	if r.errors != nil {
		return nil, errors.New("error listener already registered")
	}

	r.errors = make(chan error, 1)
	return r.errors, nil
}
