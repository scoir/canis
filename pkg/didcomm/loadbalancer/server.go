package loadbalancer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/transport"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"nhooyr.io/websocket"

	"github.com/scoir/canis/pkg/amqp"
	"github.com/scoir/canis/pkg/amqp/rabbitmq"
	"github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
)

const (
	// TODO configure ping request frequency.
	didcommPrefix = "https://didcomm.org/"
	pingFrequency = 30 * time.Second
)

var (
	supportedProtocols = []string{"didexchange", "issue-credential"}
)

type Server struct {
	wsAddr     string
	httpAddr   string
	external   string
	packager   transport.Packager
	publishers map[string]amqp.Publisher
}

type provider interface {
	Packager() transport.Packager
}

//TODO:  to make this testable, need a "PublisherProvider" that will create the publisher here so we aren't hardcoding to RabbitMQ
func New(prov provider, amqpAddr, host string, httpPort, wsPort int, external string) (*Server, error) {

	qs := make(map[string]amqp.Publisher)
	for _, queueName := range supportedProtocols {
		p, err := rabbitmq.NewPublisher(amqpAddr, queueName)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create AMQP Publisher")
		}
		qs[queueName] = p
	}

	return &Server{
		wsAddr:     fmt.Sprintf("%s:%d", host, wsPort),
		httpAddr:   fmt.Sprintf("%s:%d", host, httpPort),
		external:   external,
		packager:   prov.Packager(),
		publishers: qs,
	}, nil
}

func (r *Server) Close() error {
	for _, pub := range r.publishers {
		err := pub.Close()
		if err != nil {
			return errors.Wrap(err, "error closing publisher")
		}
	}

	return nil
}

func (r *Server) Start() {
	go r.startWS()
	go r.startHTTP()
}

func (r *Server) RegisterGRPCHandler(server *grpc.Server) {
	api.RegisterLoadbalancerServer(server, r)
}

func (r *Server) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {
}

func (r *Server) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}

func (r *Server) GetEndpoint(_ context.Context, _ *api.EndpointRequest) (*api.EndpointResponse, error) {
	return &api.EndpointResponse{Endpoint: r.external}, nil
}

func (r *Server) startWS() {
	server := &http.Server{Addr: r.wsAddr}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.processWsRequest(w, req)
	})

	log.Printf("Listening for WS DIDComm messages on %s to queue\n", r.wsAddr)
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
				log.Printf("Error reading request message: %v", err)
			}
			break
		}

		pub, err := r.publisherFromMessage(message)
		if err != nil {
			log.Printf("error typing message: %v", err)
			continue
		}

		err = pub.Publish(message, "application/json")
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
		log.Printf("connection close error: %v\n", err)
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

		pub, err := r.publisherFromMessage(body)
		if err != nil {
			log.Printf("error typing message: %v", err)
			return
		}

		err = pub.Publish(body, "application/json")
		if err != nil {
			return
		}
	})

	srv.Handler = handler

	log.Printf("Listening for HTTP DIDComm messages on %s to queue\n", r.httpAddr)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln("error listening on HTTP", err)
	}
}

func (r *Server) publisherFromMessage(message []byte) (amqp.Publisher, error) {
	unpackMsg, err := r.packager.UnpackMessage(message)
	if err != nil {
		return nil, errors.Wrap(err, "error unpacking message")
	}
	trans := &struct {
		Type string `json:"@type"`
	}{}

	err = json.Unmarshal(unpackMsg.Message, trans)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling message")
	}

	if !strings.HasPrefix(trans.Type, didcommPrefix) {
		return nil, errors.Errorf("invalid message type: %s", trans.Type)
	}

	suffix := trans.Type[len(didcommPrefix):]
	i := strings.Index(suffix, "/")
	if i == -1 {
		return nil, errors.Errorf("invalid message suffix: %s", trans.Type)
	}

	pub, ok := r.publishers[suffix[:i]]
	if !ok {
		return nil, errors.Errorf("no publisher for protocol %s", suffix[:i])
	}

	return pub, nil

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
		http.Error(w, fmt.Sprintf("Unsupported Content-type \"%s\"", ct), http.StatusUnsupportedMediaType)
		return false
	}

	return true
}
