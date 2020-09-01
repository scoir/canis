package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"
	"nhooyr.io/websocket"
)

var ch *amqp.Channel

const (
	// TODO configure ping request frequency.
	queueName     = "didcomm-msgs"
	pingFrequency = 30 * time.Second
)

func main() {
	listen()
}

func listen() {
	internalAddr := "0.0.0.0:3001"

	server := &http.Server{Addr: internalAddr}

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		processRequest(w, r)
	})

	var err error
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalln("unable to dial RMQ", err)
	}
	defer conn.Close()

	ch, err = conn.Channel()
	if err != nil {
		log.Fatalln("can't create channel", err)
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	fmt.Printf("listening for messages on %s to queue\n", internalAddr)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln("listen err: ", err)
	}
	log.Println("closed.")
}

func processRequest(w http.ResponseWriter, r *http.Request) {
	c, err := upgradeConnection(w, r)
	if err != nil {
		log.Fatalf("failed to upgrade the connection : %v", err)
		return
	}

	listener(c, false)
}

func listener(conn *websocket.Conn, outbound bool) {
	var verKeys []string

	defer klose(conn, verKeys)

	go keepConnAlive(conn, outbound, pingFrequency)

	for {
		_, message, err := conn.Read(context.Background())
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Fatalf("Error reading request message: %v", err)
			}

			break
		}

		fmt.Println("ENQUEING MSG")
		fmt.Println(string(message))
		fmt.Println("**********************************************")

		err = ch.Publish(
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        message,
			})

		//unpackMsg, err := d.packager.UnpackMessage(message)
		//if err != nil {
		//	log.Printf("failed to unpack msg: %v", err)
		//	continue
		//}

		//err = messageHandler(unpackMsg.Message, unpackMsg.ToDID, unpackMsg.FromDID)
		//if err != nil {
		//	log.Fatalf("incoming msg processing failed: %v", err)
		//}
	}
}

func upgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	c, err := Accept(w, r)
	if err != nil {
		log.Fatalf("failed to upgrade the connection : %v", err)
		return nil, err
	}

	return c, nil
}

func Accept(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	// TODO Allow user to enable InsecureSkipVerify https://github.com/hyperledger/aries-framework-go/issues/928
	return websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
		CompressionMode:    websocket.CompressionDisabled,
	})
}

func klose(conn *websocket.Conn, verKeys []string) {
	if err := conn.Close(websocket.StatusNormalClosure,
		"closing the connection"); websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		log.Printf("connection close error\n")
	}

}

func keepConnAlive(conn *websocket.Conn, outbound bool, frequency time.Duration) {
	if outbound {
		ticker := time.NewTicker(frequency)
		done := make(chan struct{})

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := conn.Ping(context.Background()); err != nil {
					log.Fatalf("websocket ping error : %v", err)

					ticker.Stop()
					done <- struct{}{}
				}
			}
		}
	}
}
