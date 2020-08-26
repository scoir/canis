/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import "google.golang.org/grpc"

func (r *Agent) RegisterGRPCHandler(_ *grpc.Server) {
	//NO-OP for now
}

func (r *Agent) GetServerOpts() []grpc.ServerOption {
	return []grpc.ServerOption{}
}
