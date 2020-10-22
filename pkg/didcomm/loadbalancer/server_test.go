package loadbalancer

import (
	"context"
	"reflect"
	"testing"

	"github.com/scoir/canis/pkg/protogen/common"
)

func TestServer_GetEndpoint(t *testing.T) {
	type fields struct {
		wsAddr string
	}
	type args struct {
		in0 context.Context
		in1 *common.EndpointRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *common.EndpointResponse
		wantErr bool
	}{
		{
			name: "with addr",
			fields: fields{
				wsAddr: "0.0.0.0:9999",
			},
			args: args{
				in0: nil,
				in1: &common.EndpointRequest{},
			},
			want: &common.EndpointResponse{
				Endpoint: "0.0.0.0:9999",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Server{
				external: tt.fields.wsAddr,
			}
			got, err := r.GetEndpoint(tt.args.in0, tt.args.in1)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEndpoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}
