/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import (
	"fmt"
)

type Endpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (r Endpoint) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
