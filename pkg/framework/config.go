/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"flag"
	"fmt"
)

var configFileName string

func init() {
	flag.StringVar(&configFileName, "config", "agent-config", "override default config file name")

}

type Config struct {
	ID       string       `mapstructure:"ID"`
	Endpoint string       `mapstructure:"endpoint"`
	VDR      Endpoint     `mapstructure:"vdr"`
	Agency   *AgentConfig `mapstructure:"agency"`
	Router   *AgentConfig `mapstructure:"router"`
	Steward  *AgentConfig `mapstructure:"steward"`
	Agent    *AgentConfig `mapstructure:"agent"`

	Datastore  DatastoreConfig `mapstructure:"datastore"`
	Execution  RuntimeConfig   `mapstructure:"execution"`
	GRPCConfig `mapstructure:",squash"`
}

type Endpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (r Endpoint) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
