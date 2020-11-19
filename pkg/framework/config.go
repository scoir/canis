/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"fmt"
)

type Endpoint struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Token    string `mapstructure:"token"`
}

func (r Endpoint) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type AMQPConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	VHost    string `yaml:"vhost"`
}

func (r *AMQPConfig) Endpoint() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", r.User, r.Password, r.Host, r.Port, r.VHost)
}
