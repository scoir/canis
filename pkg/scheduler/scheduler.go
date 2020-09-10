package scheduler

import (
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/client/canis"
	"github.com/scoir/canis/pkg/client/informer"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/runtime"
)

type Server struct {
	exec   runtime.Executor
	client *canis.Client

	storeManager *manager.DataProviderManager
	agentStore   datastore.Store
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	StorageManager() *manager.DataProviderManager
	StorageProvider() (datastore.Provider, error)
	Executor() (runtime.Executor, error)
	CanisClient() (*canis.Client, error)
}

func New(ctx provider) (*Server, error) {
	r := &Server{}
	exec, err := ctx.Executor()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access runtime executor")
	}
	r.exec = exec
	r.storeManager = ctx.StorageManager()

	storageProvider, err := ctx.StorageProvider()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access datastore")
	}

	r.agentStore, err = storageProvider.OpenStore("canis")
	if err != nil {
		return nil, err
	}

	r.client, err = ctx.CanisClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to start scheduler")
	}

	//TODO: self heal:
	//      - scan for agents not running that should be and start them
	//      - shutdown any agents that no longer exist

	return r, nil
}

func (r *Server) Run(stopCh chan struct{}) {
	inf := r.client.AgentInformer()
	inf.AddEventHandler(informer.ResourceEventHandlerFuncs{
		AddFunc: func(n interface{}) {
			log.Println("starting agent")
			a := n.(*api.Agent)
			sts, err := r.launchAgent(a)
			if err != nil {
				log.Println(errors.Wrapf(err, "launch agent failed with status: %d", sts))
			}
		},
		DeleteFunc: func(n interface{}) {
			log.Println("shutting down agent")
			a := n.(*api.Agent)
			err := r.shutdownAgent(a)
			if err != nil {
				log.Printf("shutdown agent failed: %v", err)
			}
		},
	})

	inf.Run(stopCh)
}

func (r *Server) launchAgent(agent *api.Agent) (api.Agent_Status, error) {
	pID, err := r.exec.LaunchAgent(agent.Id)
	if err != nil {
		return api.Agent_ERROR, errors.Wrap(err, "unable to launch agent")
	}

	out := &api.LaunchAgentResponse{
		Status: api.Agent_STARTING,
	}

	w, err := r.exec.Watch(pID)
	if err != nil {
		return api.Agent_ERROR, errors.Wrap(err, "error watching agent")
	}
	stopper := time.AfterFunc(time.Minute, func() {
		w.Stop()
	})
	defer stopper.Stop()

	for event := range w.ResultChan() {
		switch event.RuntimeContext.Status() {
		case datastore.Running:
			out.Status = api.Agent_RUNNING
			return out.Status, nil
		case datastore.Error:
			out.Status = api.Agent_ERROR
			return out.Status, errors.New(out.String())
		case datastore.Completed:
			out.Status = api.Agent_TERMINATED
			return out.Status, nil
		}
	}

	return out.Status, nil
}

func (r *Server) shutdownAgent(a *api.Agent) error {
	agent, err := r.agentStore.GetAgent(a.Id)
	if err != nil {
		return errors.Wrap(err, "unable to load agent to shutdown")
	}

	if agent.PID == "" {
		return errors.Wrapf(err, "agent with ID %s is not currently running", a.Id)
	}

	err = r.exec.ShutdownAgent(agent.PID)
	if err != nil {
		return errors.Wrap(err, "unable to shutdown agent")
	}

	agent.PID = ""
	err = r.agentStore.UpdateAgent(agent)
	if err != nil {
		return errors.Wrap(err, "unable to save agent after shutdown")
	}

	return nil
}
