package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/runtime"
)

type Server struct {
	exec   runtime.Executor
	client api.AdminClient

	storeManager *manager.DataProviderManager
	agentStore   datastore.Store
}

//go:generate mockery -name=provider --structname=Provider
type provider interface {
	StorageManager() *manager.DataProviderManager
	StorageProvider() (datastore.Provider, error)
	Executor() (runtime.Executor, error)
	AdminClient() (api.AdminClient, error)
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

	r.agentStore, err = storageProvider.OpenStore("Agent")
	if err != nil {
		return nil, err
	}

	r.client, err = ctx.AdminClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to start scheduler")
	}

	return r, nil
}

func (r *Server) Run() {

}

func (r *Server) launchAgent(_ context.Context, req *api.LaunchAgentRequest) (*api.LaunchAgentResponse, error) {
	agent, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to load agent to launch: %v", err)
	}

	pID, err := r.exec.LaunchAgent(agent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to launch agent: %v", err)
	}
	agent.PID = pID

	err = r.agentStore.UpdateAgent(agent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to save agent: %v", err)
	}

	out := &api.LaunchAgentResponse{
		Status: api.Agent_STARTING,
	}
	if req.Wait {
		w, err := r.exec.Watch(agent.PID)
		if err != nil {
			log.Println("error watching agent")
		}
		stopper := time.AfterFunc(time.Minute, func() {
			w.Stop()
		})
		defer stopper.Stop()

		for event := range w.ResultChan() {
			switch event.RuntimeContext.Status() {
			case datastore.Running:
				out.Status = api.Agent_RUNNING
				return out, nil
			case datastore.Error:
				out.Status = api.Agent_ERROR
				return out, nil
			case datastore.Completed:
				out.Status = api.Agent_TERMINATED
				return out, nil
			}
		}
	}

	return out, nil
}

func (r *Server) shutdownAgent(_ context.Context, req *api.ShutdownAgentRequest) (*api.ShutdownAgentResponse, error) {
	agent, err := r.agentStore.GetAgent(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to load agent to shutdown: %v", err)
	}

	if agent.PID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "agent with ID %s is not currently running", req.Id)
	}

	err = r.exec.ShutdownAgent(agent.PID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to shutdown agent: %v", err)
	}

	agent.PID = ""
	err = r.agentStore.UpdateAgent(agent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to save agent after shutdown: %v", err)
	}

	return &api.ShutdownAgentResponse{}, nil
}
