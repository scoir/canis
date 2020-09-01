/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"github.com/hyperledger/aries-framework-go/pkg/common/log"
	commtransport "github.com/hyperledger/aries-framework-go/pkg/didcomm/common/transport"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport"
)

const (
	queueName = "didcomm-msgs"
)

var logger = log.New("aries-framework/transport/amqp")

// Inbound amqp type.
type Inbound struct {
	internalAddr      string
	externalAddr      string
	conn              *amqp.Connection
	ch                *amqp.Channel
	que               amqp.Queue
	certFile, keyFile string
	packager          commtransport.Packager
	msgHandler        transport.InboundMessageHandler
}

// NewInbound creates a new WebSocket inbound transport instance.
func NewInbound(internalAddr, externalAddr, certFile, keyFile string) (*Inbound, error) {
	if internalAddr == "" {
		return nil, errors.New("websocket address is mandatory")
	}

	if externalAddr == "" {
		externalAddr = internalAddr
	}

	return &Inbound{
		certFile:     certFile,
		keyFile:      keyFile,
		internalAddr: internalAddr,
		externalAddr: externalAddr,
	}, nil
}

// Start the http(ws) server.
func (i *Inbound) Start(prov transport.Provider) error {
	if prov == nil || prov.InboundMessageHandler() == nil {
		return errors.New("creation of inbound handler failed")
	}

	var err error
	//TODO:  Use DialTLS if certFile and keyFile are not blank
	var conn *amqp.Connection
	if i.certFile != "" && i.keyFile != "" {
		config := &tls.Config{}
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(i.certFile, i.keyFile)
		conn, err = amqp.DialTLS(fmt.Sprintf("amqp://guest:guest@%s/", i.internalAddr), config)
		if err != nil {
			return errors.Wrap(err, "unable to connect to RabbitMQ")
		}
	} else {
		conn, err = amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s/", i.internalAddr))
		if err != nil {
			return errors.Wrap(err, "unable to connect to RabbitMQ")
		}
	}

	ch, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "unable to get channel")
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	i.conn = conn
	i.ch = ch
	i.que = q
	i.packager = prov.Packager()
	i.msgHandler = prov.InboundMessageHandler()

	go func() {
		if err := i.listenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("websocket server start with address [%s] failed, cause:  %s", i.externalAddr, err)
		}
	}()

	return nil
}

func (i *Inbound) listenAndServe() error {
	msgs, err := i.ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return errors.Wrap(err, "unable to consume")
	}

	for d := range msgs {
		message := d.Body
		fmt.Println("DEQUEUED MSG")
		fmt.Println(string(message))
		fmt.Println("**********************************************")
		unpackMsg, err := i.packager.UnpackMessage(message)

		if err != nil {
			logger.Errorf("failed to unpack msg: %v", err)

			continue
		}

		trans := &decorator.Transport{}

		err = json.Unmarshal(unpackMsg.Message, trans)
		if err != nil {
			logger.Errorf("unmarshal transport decorator : %v", err)
		}

		messageHandler := i.msgHandler
		err = messageHandler(unpackMsg.Message, unpackMsg.ToDID, unpackMsg.FromDID)
		if err != nil {
			logger.Errorf("incoming msg processing failed: %v", err)
		}
	}

	return nil
}

// Stop the http(ws) server.
func (i *Inbound) Stop() error {
	if err := i.ch.Close(); err != nil {
		return fmt.Errorf("channel shutdown failed: %w", err)
	}

	if err := i.conn.Close(); err != nil {
		return fmt.Errorf("connection shutdown failed: %w", err)
	}

	return nil
}

// Endpoint provides the http(ws) connection details.
func (i *Inbound) Endpoint() string {
	return i.externalAddr
}
