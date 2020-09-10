/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport"
	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"

	commontransport "github.com/hyperledger/aries-framework-go/pkg/didcomm/common/transport"
	mockpackager "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/packager"
)

func _TestInboundTransport(t *testing.T) {
	t.Run("test inbound transport - with host/port", func(t *testing.T) {
		port := ":" + strconv.Itoa(GetRandomPort(5))
		externalAddr := "http://example.com" + port
		inbound, err := NewInbound("localhost"+port, externalAddr, "queue", "", "")
		require.NoError(t, err)
		require.Equal(t, externalAddr, inbound.Endpoint())
	})

	t.Run("test inbound transport - with host/port, no external address", func(t *testing.T) {
		internalAddr := "example.com" + ":" + strconv.Itoa(GetRandomPort(5))
		inbound, err := NewInbound(internalAddr, "", "queue", "", "")
		require.NoError(t, err)
		require.Equal(t, internalAddr, inbound.Endpoint())
	})

	t.Run("test inbound transport - without host/port", func(t *testing.T) {
		inbound, err := NewInbound(":"+strconv.Itoa(GetRandomPort(5)), "", "queue", "", "")
		require.NoError(t, err)
		require.NotEmpty(t, inbound)
		mockPackager := &mockpackager.Packager{UnpackValue: &commontransport.Envelope{Message: []byte("data")}}
		err = inbound.Start(&mockProvider{packagerValue: mockPackager})
		require.NoError(t, err)

		err = inbound.Stop()
		require.NoError(t, err)
	})

	t.Run("test inbound transport - nil context", func(t *testing.T) {
		inbound, err := NewInbound(":"+strconv.Itoa(GetRandomPort(5)), "", "queue", "", "")
		require.NoError(t, err)
		require.NotEmpty(t, inbound)

		err = inbound.Start(nil)
		require.Error(t, err)
	})

	t.Run("test inbound transport - invalid TLS", func(t *testing.T) {
		svc, err := NewInbound(":0", "", "invalid", "invalid", "invalid")
		require.NoError(t, err)

		err = svc.listenAndServe()
		require.Error(t, err)
		require.Contains(t, err.Error(), "open invalid: no such file or directory")
	})

	t.Run("test inbound transport - invalid port number", func(t *testing.T) {
		_, err := NewInbound("", "", "", "", "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "websocket address is mandatory")
	})
}

func _TestInboundDataProcessing(t *testing.T) {
	t.Run("test inbound transport - multiple invocation with same client", func(t *testing.T) {
		port := ":" + strconv.Itoa(GetRandomPort(5))

		// initiate inbound with port
		inbound, err := NewInbound(port, "", "", "", "")
		require.NoError(t, err)
		require.NotEmpty(t, inbound)

		// start server
		mockPackager := &mockpackager.Packager{UnpackValue: &commontransport.Envelope{Message: []byte("valid-data")}}
		err = inbound.Start(&mockProvider{packagerValue: mockPackager})
		require.NoError(t, err)

		// create ws client
		client, cleanup := websocketClient(t, port)
		defer cleanup()

		ctx := context.Background()

		for i := 1; i <= 5; i++ {
			err = client.Write(ctx, websocket.MessageText, []byte("random"))
			require.NoError(t, err)
		}
	})

	t.Run("test inbound transport - unpacking error", func(t *testing.T) {
		port := ":" + strconv.Itoa(GetRandomPort(5))

		// initiate inbound with port
		inbound, err := NewInbound(port, "", "", "", "")
		require.NoError(t, err)
		require.NotEmpty(t, inbound)

		// start server
		mockPackager := &mockpackager.Packager{UnpackErr: errors.New("error unpacking")}
		err = inbound.Start(&mockProvider{packagerValue: mockPackager})
		require.NoError(t, err)

		// create ws client
		client, cleanup := websocketClient(t, port)
		defer cleanup()

		ctx := context.Background()

		err = client.Write(ctx, websocket.MessageText, []byte(""))
		require.NoError(t, err)
	})

	t.Run("test inbound transport - message handler error", func(t *testing.T) {
		port := ":" + strconv.Itoa(GetRandomPort(5))

		// initiate inbound with port
		inbound, err := NewInbound(port, "", "", "", "")
		require.NoError(t, err)
		require.NotEmpty(t, inbound)

		// start server
		mockPackager := &mockpackager.Packager{UnpackValue: &commontransport.Envelope{Message: []byte("invalid-data")}}
		err = inbound.Start(&mockProvider{packagerValue: mockPackager})
		require.NoError(t, err)

		// create ws client
		client, cleanup := websocketClient(t, port)
		defer cleanup()

		ctx := context.Background()

		err = client.Write(ctx, websocket.MessageText, []byte(""))
		require.NoError(t, err)
	})

	t.Run("test inbound transport - client close error", func(t *testing.T) {
		port := ":" + strconv.Itoa(GetRandomPort(5))

		// initiate inbound with port
		inbound, err := NewInbound(port, "", "", "", "")
		require.NoError(t, err)
		require.NotEmpty(t, inbound)

		// start server
		mockPackager := &mockpackager.Packager{}
		err = inbound.Start(&mockProvider{packagerValue: mockPackager})
		require.NoError(t, err)

		// create ws client
		client, _ := websocketClient(t, port)

		err = client.Close(websocket.StatusInternalError, "abnormal closure")
		require.NoError(t, err)
	})
}

func websocketClient(t *testing.T, port string) (*websocket.Conn, func()) {
	require.NoError(t, VerifyListener("localhost"+port, time.Second))

	u := url.URL{Scheme: "ws", Host: "localhost" + port, Path: ""}
	c, resp, err := websocket.Dial(context.Background(), u.String(), nil) // nolint - bodyclose (library closes the body)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	return c, func() {
		require.NoError(t, c.Close(websocket.StatusNormalClosure, "closing the connection"))
	}
}

type mockProvider struct {
	packagerValue commontransport.Packager
}

func (p *mockProvider) InboundMessageHandler() transport.InboundMessageHandler {
	return func(message []byte, myDID, theirDID string) error {
		logger.Infof("message received is %s", string(message))

		if string(message) == "invalid-data" {
			return errors.New("error")
		}

		return nil
	}
}

func (p *mockProvider) Packager() commontransport.Packager {
	return p.packagerValue
}

func (p *mockProvider) AriesFrameworkID() string {
	return uuid.New().String()
}

func GetRandomPort(n int) int {
	for ; n > 0; n-- {
		port, err := getRandomPort()
		if err != nil {
			continue
		}

		return port
	}
	panic("cannot acquire the random port")
}

func getRandomPort() (int, error) {
	const network = "tcp"

	addr, err := net.ResolveTCPAddr(network, "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP(network, addr)
	if err != nil {
		return 0, err
	}

	err = listener.Close()
	if err != nil {
		return 0, err
	}

	return listener.Addr().(*net.TCPAddr).Port, nil
}

// VerifyListener verifies if the host/port is available for listening.
func VerifyListener(host string, d time.Duration) error {
	timeout := time.After(d)

	for {
		select {
		case <-timeout:
			return errors.New("timeout: server is not available")
		default:
			conn, err := net.Dial("tcp", host)
			if err != nil {
				continue
			}

			return conn.Close()
		}
	}
}
