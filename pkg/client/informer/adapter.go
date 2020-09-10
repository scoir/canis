package informer

import (
	"context"
	"io"
	"log"

	"github.com/cenkalti/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/util"
)

type Update struct {
	Old interface{}
	New interface{}
}

type StreamAdapter interface {
	AddCh() chan interface{}
	UpdateCh() chan Update
	DeleteCh() chan interface{}
	Close()
}

type agentStreamAdapter struct {
	client api.AdminClient
	cancel context.CancelFunc
	addCh  chan interface{}
	updCh  chan Update
	delCh  chan interface{}
}

func NewAgentStreamAdapter(s api.AdminClient) StreamAdapter {
	out := &agentStreamAdapter{
		client: s,
		addCh:  make(chan interface{}),
		updCh:  make(chan Update),
		delCh:  make(chan interface{}),
	}
	go func() {
		log.Print("watching agent events...")
		err := backoff.RetryNotify(out.watch, backoff.NewExponentialBackOff(), util.Logger)
		log.Println(err)
	}()
	return out
}

func (r *agentStreamAdapter) AddCh() chan interface{} {
	return r.addCh
}

func (r *agentStreamAdapter) UpdateCh() chan Update {
	return r.updCh
}

func (r *agentStreamAdapter) DeleteCh() chan interface{} {
	return r.delCh
}

func (r *agentStreamAdapter) Close() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *agentStreamAdapter) watch() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	stream, err := r.client.WatchAgents(ctx, &api.WatchRequest{})
	if err != nil {
		return err
	}

	r.cancel = cancelFunc

	log.Println("stream successfully created")
	for {
		evt, err := stream.Recv()
		if err == io.EOF || status.Code(err) == codes.Canceled {
			return nil
		}

		if err != nil {
			return err
		}

		switch evt.Type {
		case api.AgentEvent_ADD:
			r.addCh <- evt.New
		case api.AgentEvent_UPDATE:
			r.updCh <- Update{Old: evt.Old, New: evt.New}
		case api.AgentEvent_DELETE:
			r.delCh <- evt.Old
		}

	}

}
