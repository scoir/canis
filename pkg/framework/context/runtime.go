/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/runtime"
	"github.com/scoir/canis/pkg/runtime/docker"
	"github.com/scoir/canis/pkg/runtime/kubernetes"
)

const (
	executionKey = "execution"
)

type RuntimeConfig struct {
	Runtime    string             `mapstructure:"runtime"`
	Kubernetes *kubernetes.Config `mapstructure:"kubernetes"`
	Docker     *docker.Config     `mapstructure:"docker"`
}

func (r *Provider) Executor() (runtime.Executor, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.exec != nil {
		return r.exec, nil
	}

	rtc := &RuntimeConfig{}
	err := r.vp.UnmarshalKey(executionKey, rtc)
	if err != nil {
		return nil, errors.Wrap(err, "execution environment is not correctly configured")
	}
	switch rtc.Runtime {
	case "kubernetes":
		r.exec, err = r.loadK8s()
	case "docker":
		r.exec, err = r.loadDocker(rtc.Docker)
	default:
		return nil, errors.New("no known execution environment is configured")
	}

	return r.exec, errors.Wrap(err, "unable to launch runtime from config")

}

func (r *Provider) loadK8s() (runtime.Executor, error) {
	return nil, errors.New("not implemented")
}

func (r *Provider) loadDocker(dc *docker.Config) (runtime.Executor, error) {
	if dc == nil {
		return nil, errors.New("docker execution environment not properly configured")
	}
	d, err := docker.New(r, dc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create docker execution environment")
	}
	return d, nil
}
