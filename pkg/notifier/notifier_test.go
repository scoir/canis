package notifier

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	samqp "github.com/streadway/amqp"
	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/amqp"
	lmocks "github.com/scoir/canis/pkg/amqp/mocks"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/mocks"
)

type mockProvider struct {
	ds       *mocks.Store
	listener *lmocks.Listener
}

func (m mockProvider) GetDatastore() datastore.Store {
	return m.ds
}

func (m mockProvider) GetAMQPListener(queue string) amqp.Listener {
	return m.listener
}

func TestServer_Start(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		eventCh := make(chan []byte, 1)
		testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			eventMessage, err := ioutil.ReadAll(req.Body)
			require.NoError(t, err)
			w.WriteHeader(http.StatusNoContent)
			eventCh <- eventMessage
		}))
		note := Notification{
			Topic:     "test-topic",
			Event:     "testing",
			EventData: map[string]interface{}{"test": 123},
		}
		webhooks := []*datastore.Webhook{
			{
				Type: "test-topic",
				URL:  testSrv.URL,
			},
		}

		msgs := make(chan samqp.Delivery, 1)
		prov.listener.On("Listen").Return((<-chan samqp.Delivery)(msgs), nil)
		prov.ds.On("ListWebhooks", note.Topic).Return(webhooks, nil)

		go func() {
			err := target.Start()
			require.NoError(t, err)
		}()

		d, err := json.Marshal(note)
		msgs <- samqp.Delivery{
			ContentType: "application/json",
			Body:        d,
		}

		eventMessage := <-eventCh
		m := map[string]interface{}{}
		_ = json.Unmarshal(eventMessage, &m)
		require.Equal(t, "testing", m["event"])
		require.Equal(t, map[string]interface{}{"test": float64(123)}, m["message"])

	})

	t.Run("bad response", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("bad err"))
		}))
		note := Notification{
			Topic:     "test-topic",
			Event:     "testing",
			EventData: map[string]interface{}{"test": 123},
		}
		webhooks := []*datastore.Webhook{
			{
				Type: "test-topic",
				URL:  testSrv.URL,
			},
		}

		msgs := make(chan samqp.Delivery, 1)
		prov.listener.On("Listen").Return((<-chan samqp.Delivery)(msgs), nil)
		prov.ds.On("ListWebhooks", note.Topic).Return(webhooks, nil)

		errCh, err := target.Errors()
		go func() {
			err := target.Start()
			require.NoError(t, err)
		}()

		d, err := json.Marshal(note)
		msgs <- samqp.Delivery{
			ContentType: "application/json",
			Body:        d,
		}

		err = <-errCh
		require.Error(t, err)

	})

	t.Run("webhook failure", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		note := Notification{
			Topic:     "test-topic",
			Event:     "testing",
			EventData: map[string]interface{}{"test": 123},
		}

		msgs := make(chan samqp.Delivery, 1)
		prov.listener.On("Listen").Return((<-chan samqp.Delivery)(msgs), nil)
		prov.ds.On("ListWebhooks", note.Topic).Return(nil, errors.New("not found"))

		errCh, err := target.Errors()
		go func() {
			err := target.Start()
			require.NoError(t, err)
		}()

		d, err := json.Marshal(note)
		msgs <- samqp.Delivery{
			ContentType: "application/json",
			Body:        d,
		}

		err = <-errCh
		require.Error(t, err)
	})

	t.Run("invalid message", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		msgs := make(chan samqp.Delivery, 1)
		prov.listener.On("Listen").Return((<-chan samqp.Delivery)(msgs), nil)

		errCh, err := target.Errors()
		require.NoError(t, err)
		go func() {
			err := target.Start()
			require.NoError(t, err)
		}()

		msgs <- samqp.Delivery{
			ContentType: "application/json",
			Body:        []byte(`{`),
		}

		err = <-errCh
		require.Error(t, err)

	})

	t.Run("listener error", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		prov.listener.On("Listen").Return(nil, errors.New("BOOM"))
		err = target.Start()
		require.Error(t, err)
	})

	t.Run("messages closed", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		msgs := make(chan samqp.Delivery, 1)
		prov.listener.On("Listen").Return((<-chan samqp.Delivery)(msgs), nil)

		errCh := make(chan error, 1)
		go func() {
			err := target.Start()
			errCh <- err
		}()

		close(msgs)

		startErr := <-errCh
		require.Equal(t, "notification messages closed", startErr.Error())

	})

}

func TestServer_Errors(t *testing.T) {
	t.Run("only one error listener", func(t *testing.T) {
		prov := &mockProvider{
			ds:       &mocks.Store{},
			listener: &lmocks.Listener{},
		}

		target, err := New(prov)
		require.NoError(t, err)

		_, err = target.Errors()
		require.NoError(t, err)

		_, err = target.Errors()
		require.Error(t, err)
	})
}
