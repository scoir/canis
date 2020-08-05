/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package steward

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/scoir/canis/pkg/proxy"
)

func (r *Steward) GetServerOpts() []grpc.ServerOption {
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		//// Make sure we never forward internal services.
		if strings.HasPrefix(fullMethodName, "/steward.") {
			return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// Decide on which backend to dial
			if val, exists := md["college-agent-id"]; exists {
				// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
				conn, err := grpc.DialContext(ctx, fmt.Sprintf("college-agent-%s:7777", val[0]),
					grpc.WithCodec(proxy.Codec()),
					grpc.WithInsecure(),
				)
				return ctx, conn, err
			} else if val, exists := md["highschool-agent-id"]; exists {
				conn, err := grpc.DialContext(ctx, fmt.Sprintf("highschool-agent-%s:7777", val[0]),
					grpc.WithCodec(proxy.Codec()),
					grpc.WithInsecure(),
				)

				return ctx, conn, err
			}
		}
		return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
	}

	return []grpc.ServerOption{
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
	}
}
