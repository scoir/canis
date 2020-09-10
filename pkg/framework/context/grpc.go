/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	api "github.com/scoir/canis/pkg/apiserver/api"
	"github.com/scoir/canis/pkg/framework"
)

const (
	grpcKey     = "grpc"
	bridgeKey   = "grpcbridge"
	apiEndpoint = "api.grpc"
)

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	if !r.vp.IsSet(grpcKey) {
		return nil, errors.New("grpc is not properly configured")
	}

	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey(grpcKey, ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	if !r.vp.IsSet(bridgeKey) {
		return nil, errors.New("grpc bridge is not properly configured")
	}

	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey(bridgeKey, ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc bridge is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetAPIAdminClient() (api.AdminClient, error) {
	if !r.vp.IsSet(apiEndpoint) {
		return nil, errors.New("steward client is not properly configured")
	}

	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey(apiEndpoint, ep)
	if err != nil {
		return nil, errors.Wrap(err, "steward client is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for steward client")
	}
	cl := api.NewAdminClient(cc)
	return cl, nil
}
