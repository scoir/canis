/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	api "github.com/scoir/canis/pkg/apiserver/api/protogen"
)

func (r *Provider) GetAPIAdminClient() (api.AdminClient, error) {
	ep, err := r.conf.Endpoint("api.grpc")
	if err != nil {
		return nil, err
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for api client")
	}

	cl := api.NewAdminClient(cc)
	return cl, nil
}
