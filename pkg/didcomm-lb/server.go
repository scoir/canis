package lb

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"nhooyr.io/websocket"
)

const (
	// TODO configure ping request frequency.
	queueName     = "didcomm-msgs"
	pingFrequency = 30 * time.Second
)

type Server struct {
	wsAddr   string
	httpAddr string
	conn     *amqp.Connection
	ch       *amqp.Channel
	queue    amqp.Queue
}

func New(amqpAddr, host string, httpPort, wsPort int) (*Server, error) {

	var err error
	conn, err := amqp.Dial(amqpAddr)
	if err != nil {
		log.Fatalln("unable to dial RMQ", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create an AMQP channel")
	}

	queue, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to declare AMQP queue")
	}

	return &Server{
		wsAddr:   fmt.Sprintf("%s:%d", host, wsPort),
		httpAddr: fmt.Sprintf("%s:%d", host, httpPort),
		conn:     conn,
		ch:       ch,
		queue:    queue,
	}, nil
}

func (r *Server) Close() error {
	err := r.conn.Close()
	if err != nil {
		return errors.Wrap(err, "unable to ")
	}
	return nil
}

func (r *Server) Start() {
	go r.startWS()
	go r.startHTTP()
}

func (r *Server) startWS() {
	server := &http.Server{Addr: r.wsAddr}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.processWsRequest(w, req)
	})

	fmt.Printf("listening for WS DIDComm messages on %s to queue\n", r.wsAddr)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalln("listen err: ", err)
		}
		log.Println("closed.")
	}()

}
func (r *Server) processWsRequest(w http.ResponseWriter, req *http.Request) {
	c, err := upgradeConnection(w, req)
	if err != nil {
		log.Fatalf("failed to upgrade the connection : %v", err)
		return
	}

	r.listener(c, false)
}

func (r *Server) listener(conn *websocket.Conn, outbound bool) {
	var verKeys []string

	defer closeWs(conn, verKeys)

	go keepConnAlive(conn, outbound, pingFrequency)

	for {
		_, message, err := conn.Read(context.Background())
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Fatalf("Error reading request message: %v", err)
			}

			break
		}

		err = r.ch.Publish(
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        message,
			})

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

func closeWs(conn *websocket.Conn, verKeys []string) {
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

func (r *Server) startHTTP() {

	srv := &http.Server{Addr: r.httpAddr}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if valid := validateHTTPMethod(w, req); !valid {
			return
		}

		if valid := validatePayload(req, w); !valid {
			return
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read payload", http.StatusInternalServerError)
			return
		}

		err = r.ch.Publish(
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			return
		}
	})

	srv.Handler = handler

	fmt.Printf("listening for HTTP DIDComm messages on %s to queue\n", r.httpAddr)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("error listening on HTTP", err)
	}
}

func validatePayload(r *http.Request, w http.ResponseWriter) bool {
	if r.ContentLength == 0 { // empty payload should not be accepted
		http.Error(w, "Empty payload", http.StatusBadRequest)
		return false
	}

	return true
}

func validateHTTPMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		http.Error(w, "HTTP Method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	ct := r.Header.Get("Content-type")
	if ct != "application/didcomm-envelope-enc" {
		log.Println("hihihihi")
		http.Error(w, fmt.Sprintf("Unsupported Content-type \"%s\"", ct), http.StatusUnsupportedMediaType)
		return false
	}

	return true
}
