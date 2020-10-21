package informer

import (
	"context"
	"log"

	"github.com/cenkalti/backoff"

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
	return nil
}
