/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/runtime"
)

const (
	APIServerName              = "canis-apiserver"
	APIServerConfig            = "%s/steward_config.yaml"
	APIServerContainerName     = "canis-apiserver"
	APIServerInitContainerName = "init-canis-apiserver"
	CanisImage                 = "canis/canis:latest"

	AgentContainerName = "canis_agent_%s"
	AgentConfig        = "%s/agent_%s_config.yaml"
)

type Config struct {
	HomeDir string `mapstructure:"home"`
}

type Executor struct {
	home     string
	ctx      provider
	dockercl *client.Client
}

type provider interface {
	GetAgentConfig(agentID string) (map[string]interface{}, error)
}

func New(conf *Config) (runtime.Executor, error) {
	r := &Executor{
		home: conf.HomeDir,
	}

	var err error
	if r.home == "" {
		r.home = "/tmp"
	}
	r.dockercl, err = client.NewEnvClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to launch docker")
	}

	return r, nil
}

type dockerProc struct {
	pid    string
	name   string
	config string
	status datastore.StatusType
	dur    time.Duration
}

func (r *dockerProc) ID() string {
	return r.pid
}

func (r *dockerProc) Name() string {
	return r.name
}

func (r *dockerProc) Config() string {
	return r.config
}

func (r *dockerProc) Status() datastore.StatusType {
	return r.status
}

func (r *dockerProc) Time() time.Duration {
	return r.dur
}

func (r *Executor) AgentPS() []runtime.Process {
	return []runtime.Process{}
}

func (r *Executor) PS() []runtime.Process {
	out := make([]runtime.Process, 0)
	stewardConfigFile := fmt.Sprintf(APIServerConfig, r.home)
	proc := &dockerProc{
		name:   APIServerName,
		config: stewardConfigFile,
	}
	steward, err := r.getRunningConainer(APIServerContainerName)
	if err == nil {
		proc.pid = steward.ID[:12]
		proc.status = r.processStatus(steward)
		proc.dur = time.Now().Sub(time.Unix(steward.Created, 0))
	}

	out = append(out, proc)

	return out
}

func (r *Executor) GetAgentConfig(_ string) (map[string]interface{}, error) {
	//TODO:  Build Agent config somehow, but not this way
	return map[string]interface{}{}, nil
}

func (r *Executor) LaunchAgent(agentID string) (string, error) {
	ctx := context.Background()
	agentContainerName := fmt.Sprintf(AgentContainerName, agentID)

	agent, err := r.getRunningConainer(agentContainerName)
	if err == nil {
		state := r.processStatus(agent)
		if state == datastore.Running {
			return "", errors.Errorf("Agent %s is already running", agentContainerName)
		}
	}

	_ = r.removeContainer(agentContainerName)

	am, err := r.GetAgentConfig(agentID)
	if err != nil {
		return "", errors.Wrap(err, "unexpected error generating agent configuration")
	}
	agentConfigFileData, err := yaml.Marshal(am)
	if err != nil {
		return "", errors.Wrap(err, "unexpected yaml marshal error")
	}
	agentConfigFile := fmt.Sprintf(AgentConfig, "/etc/canis", agentID)
	err = ioutil.WriteFile(agentConfigFile, agentConfigFileData, 0644)
	if err != nil {
		return "", errors.Wrap(err, "unable to write config for agent")
	}

	p1, _ := r.FreePort()
	p2, _ := r.FreePort()
	host := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8888/tcp": []nat.PortBinding{{"0.0.0.0", fmt.Sprintf("%d", p1)}},
			"8889/tcp": []nat.PortBinding{{"0.0.0.0", fmt.Sprintf("%d", p2)}},
		},
		AutoRemove: false,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: r.home,
				Target: "/etc/canis",
			},
		}}

	nt := &network.NetworkingConfig{}

	conf := &container.Config{
		Image: CanisImage,
		Cmd: []string{
			"agent", "start",
			"--config", agentConfigFile,
			"--id", agentID,
		},

		ExposedPorts: nat.PortSet{
			"8888/tcp": struct{}{},
			"8889/tcp": struct{}{},
		},
	}

	resp, err := r.dockercl.ContainerCreate(ctx, conf, host, nt, agentContainerName)
	if err != nil {
		return "", errors.Wrap(err, "unable to create container for agenta")
	}

	if err := r.dockercl.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", errors.Wrap(err, "unable to start agent")
	}

	return resp.ID[:12], nil
}

func (r *Executor) Status(pID string) (runtime.Process, error) {
	panic("implement me")
}

func (r *Executor) ShutdownAgent(agentID string) error {
	panic("implement me")
}

func (r *Executor) Watch(pID string) (runtime.Watcher, error) {
	panic("implement me")
}

func (r *Executor) StreamLogs(pID string) (io.ReadCloser, error) {
	panic("implement me")
}

func (r *Executor) Describe() {
	panic("implement me")
}

func (r *Executor) getRunningConainer(name string) (*types.Container, error) {
	args := filters.NewArgs()
	args.Add("name", name)
	containers, err := r.dockercl.ContainerList(context.Background(), types.ContainerListOptions{Filters: args})
	if err != nil || len(containers) == 0 {
		return nil, errors.New("running container not found")
	}

	return &containers[0], nil
}

func (r *Executor) processStatus(container *types.Container) datastore.StatusType {

	switch container.State {
	case "running":
		return datastore.Running
	case "exited":
		return datastore.Completed
	case "error":
		return datastore.Error
	default:
		return datastore.NotStarted
	}
}

func (r *Executor) removeContainer(name string) error {
	ctx := context.Background()
	args := filters.NewArgs()
	args.Add("name", name)
	containers, err := r.dockercl.ContainerList(ctx, types.ContainerListOptions{Filters: args, All: true})
	if err != nil || len(containers) == 0 {
		return nil
	}

	err = r.dockercl.ContainerRemove(ctx, containers[0].ID, types.ContainerRemoveOptions{})
	return errors.Wrap(err, "unexpected error trying to remove container")
}

func (r *Executor) FreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
