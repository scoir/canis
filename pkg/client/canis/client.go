package canis

import (
	"log"
	"sync"

	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/client/informer"
)

type Client struct {
	lock          sync.Mutex
	client        api.AdminClient
	agentInformer *informer.SharedResourceInformer
}

func New(endpoint string) *Client {
	cc, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalln("can't connect", err)
	}
	r := &Client{
		client: api.NewAdminClient(cc),
	}

	return r
}

func (r *Client) Close() {
}

func (r *Client) AgentInformer() informer.ResourceInformer {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.agentInformer == nil {
		adapt := informer.NewAgentStreamAdapter(r.client)
		r.agentInformer = informer.NewSharedResourceInformer(adapt)
	}

	return r.agentInformer
}
