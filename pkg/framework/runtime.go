package framework

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/runtime"
	"github.com/scoir/canis/pkg/runtime/docker"
	"github.com/scoir/canis/pkg/runtime/kubernetes"
)

type RuntimeConfig struct {
	Runtime    string             `mapstructure:"runtime"`
	Kubernetes *kubernetes.Config `mapstructure:"kubernetes"`
	Proc       *docker.Config     `mapstructure:"proc"`

	lock sync.Mutex
	exec runtime.Executor
}

func (r *RuntimeConfig) Executor() (runtime.Executor, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.exec != nil {
		return r.exec, nil
	}

	var err error
	switch r.Runtime {
	case "kubernetes":
		r.exec, err = r.loadK8s()
	case "proc":
		r.exec, err = r.loadDocker()
	case "docker":
	}

	return r.exec, errors.Wrap(err, "unable to launch runtime from config")

}

func (r *RuntimeConfig) loadK8s() (runtime.Executor, error) {

	return nil, nil
}

func (r *RuntimeConfig) loadDocker() (runtime.Executor, error) {
	d, err := docker.New(r.Proc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create docker execution environment")
	}
	return d, nil
}
