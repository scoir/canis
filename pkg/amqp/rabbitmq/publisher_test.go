package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublisher_Publish(t *testing.T) {
	t.Run("publish and listen", func(t *testing.T) {
		addy := "amqp://rabbit:5672/"
		queue := "test-queue"
		publisher, err := NewPublisher(addy, queue)
		require.NoError(t, err)
		listener, err := NewListener(addy, queue)
		require.NoError(t, err)

		msgCh := make(chan []byte, 1)
		go func() {
			ch, err := listener.Listen()
			require.NoError(t, err)

			incoming := <-ch
			msgCh <- incoming.Body
		}()

		err = publisher.Publish([]byte("{}"), "application/json")
		require.NoError(t, err)

		err = publisher.Close()
		require.NoError(t, err)

		result := <-msgCh
		require.Equal(t, []byte("{}"), result)

		err = listener.Close()
		require.NoError(t, err)

	})

	t.Run("bad address publisher", func(t *testing.T) {
		addy := "amqp://localhost:9999/"
		queue := "test-queue"
		publisher, err := NewPublisher(addy, queue)
		require.Error(t, err)
		require.Nil(t, publisher)
	})

	t.Run("bad address listener", func(t *testing.T) {
		addy := "amqp://localhost:9999/"
		queue := "test-queue"
		listener, err := NewListener(addy, queue)
		require.Error(t, err)
		require.Nil(t, listener)
	})

}
