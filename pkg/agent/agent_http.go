/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package agent

import (
	"errors"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func (r *Agent) RegisterGRPCGateway(_ *runtime.ServeMux, _ string, _ ...grpc.DialOption) {
	//NO-OP for now
}

func (r *Agent) APISpec() (http.HandlerFunc, error) {
	return nil, errors.New("not implemented")
}
