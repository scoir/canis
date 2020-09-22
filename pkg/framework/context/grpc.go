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
	apiEndpoint = "api.grpc"
)

func (r *Provider) GetAPIAdminClient() (api.AdminClient, error) {
	if !r.vp.IsSet(apiEndpoint) {
		return nil, errors.New("api client is not properly configured")
	}

	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey(apiEndpoint, ep)
	if err != nil {
		return nil, errors.Wrap(err, "api client is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for api client")
	}
	cl := api.NewAdminClient(cc)
	return cl, nil
}
