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
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/runtime"
)

const (
	StewardName          = "steward"
	StewardConfig        = "%s/steward_config.yaml"
	StewardContainerName = "canis_steward"
	StewardImage         = "canis/steward:latest"
)

type Config struct {
	HomeDir string `mapstructure:"home"`
}

type Executor struct {
	home string
	cli  *client.Client
}

func New(conf *Config) (runtime.Executor, error) {
	r := &Executor{
		home: conf.HomeDir,
	}

	var err error
	if r.home == "" {
		r.home = "/tmp"
	}
	r.cli, err = client.NewEnvClient()
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
	stewardConfigFile := fmt.Sprintf(StewardConfig, r.home)
	proc := &dockerProc{
		name:   StewardName,
		config: stewardConfigFile,
	}
	steward, err := r.getRunningConainer(StewardContainerName)
	if err == nil {
		proc.pid = steward.ID[:12]
		proc.status = r.processStatus(steward)
		proc.dur = time.Now().Sub(time.Unix(steward.Created, 0))
	}

	out = append(out, proc)

	return out
}

func (r *Executor) LaunchSteward(configFileData []byte) (string, error) {

	ctx := context.Background()

	steward, err := r.getRunningConainer(StewardContainerName)
	if err == nil {
		state := r.processStatus(steward)
		if state == datastore.Running {
			return "", errors.New("Steward is already running")
		}
	}

	_ = r.removeContainer(StewardContainerName)

	stewardConfigFile := fmt.Sprintf(StewardConfig, r.home)
	err = ioutil.WriteFile(stewardConfigFile, configFileData, 0644)
	if err != nil {
		return "", errors.Wrap(err, "unable to write config for steward")
	}

	host := &container.HostConfig{
		PortBindings: nat.PortMap{
			"7778/tcp": []nat.PortBinding{{"0.0.0.0", "7778"}},
			"7779/tcp": []nat.PortBinding{{"0.0.0.0", "7779"}},
		},
		AutoRemove: false,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: r.home,
				Target: "/etc/canis",
			},
		}}

	net := &network.NetworkingConfig{}

	conf := &container.Config{
		Image: StewardImage,
		Cmd: []string{
			"steward", "start",
		},
		ExposedPorts: nat.PortSet{
			"7778/tcp": struct{}{},
			"7779/tcp": struct{}{},
		},
	}

	resp, err := r.cli.ContainerCreate(ctx, conf, host, net, StewardContainerName)
	if err != nil {
		return "", errors.Wrap(err, "unable to create container for steward")
	}

	if err := r.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", errors.Wrap(err, "unable to start steward")
	}

	return resp.ID[:12], nil
}

func (r *Executor) ShutdownSteward() error {

	steward, err := r.getRunningConainer(StewardContainerName)
	if err == nil {
		ctx := context.Background()
		var timeout = time.Minute
		err := r.cli.ContainerStop(ctx, steward.ID, &timeout)
		if err != nil {
			return errors.Wrap(err, "unable to stop steward")
		}
	}
	_ = r.removeContainer(StewardContainerName)
	return nil

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

func (r *Executor) Describe() {
	panic("implement me")
}

func (r *Executor) getRunningConainer(name string) (*types.Container, error) {
	args := filters.NewArgs()
	args.Add("name", name)
	containers, err := r.cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: args})
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
	containers, err := r.cli.ContainerList(ctx, types.ContainerListOptions{Filters: args, All: true})
	if err != nil || len(containers) == 0 {
		return nil
	}

	err = r.cli.ContainerRemove(ctx, containers[0].ID, types.ContainerRemoveOptions{})
	return errors.Wrap(err, "unexpected error trying to remove container")
}
