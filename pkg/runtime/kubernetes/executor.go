/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package kubernetes

import (
	"io"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/runtime"
)

type Config struct {
	KubeConfig    string `yaml:"kubeConfig"`
	Namespace     string `yaml:"namespace"`
	FQDN          string `yaml:"FQDN"`
	ImageRegistry string `yaml:"imageRegistry"`
}

type Executor struct {
}

func (r *Executor) InitSteward(seed string, d []byte) (string, error) {
	panic("implement me")
}

func (r *Executor) ShutdownSteward() error {
	panic("implement me")
}

func (r *Executor) AgentPS() []runtime.Process {
	panic("implement me")
}

func (r *Executor) PS() []runtime.Process {
	panic("implement me")
}

func New(conf *Config) runtime.Executor {
	return &Executor{}
}

func (r *Executor) LaunchSteward(conf []byte) (string, error) {
	panic("implement me")
}

func (r *Executor) LaunchAgent(agent *datastore.Agent) (string, error) {
	panic("implement me")
}

func (r *Executor) Status(pID string) (runtime.Process, error) {
	panic("implement me")
}

func (r *Executor) ShutdownAgent(pID string) error {
	panic("implement me")
}

func (r *Executor) Watch(pID string) (runtime.Watcher, error) {
	panic("implement me")
}

func (r *Executor) StreamLogs(pID string) (io.ReadCloser, error) {
	panic("implement me")
}

func GetClientSet(kubeconfig, namespace string) *framework.Clientset {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalln("error building in cluster config", err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalln("error building local config from file", err)
		}
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return &framework.Clientset{Clientset: cs, Namespace: namespace}
}

func GetClientSetWithConfig(c *Config) *framework.Clientset {
	return GetClientSet(c.KubeConfig, c.Namespace)
}

func (r *Executor) Describe() {
	panic("implement me")
}
