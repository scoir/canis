package indy

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/scoir/canis/pkg/mock/vdri/indy"
)

func TestVDRI_Read(t *testing.T) {
	type fields struct {
		methodName string
		client     indyClient
	}
	type args struct {
		did  string
		opts []vdri.ResolveOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(t require.TestingT, object interface{}, msgAndArgs ...interface{})
		wantErr bool
	}{
		{
			name:    "invalid did string",
			fields:  fields{},
			args:    args{"dXd:invalid", []vdri.ResolveOpts{}},
			want:    require.Nil,
			wantErr: true,
		},
		{
			name: "invalid method name for this VDRI",
			fields: fields{
				methodName: "sov",
			},
			args:    args{"did:peer:abc123", []vdri.ResolveOpts{}},
			want:    require.Nil,
			wantErr: true,
		},
		{
			name: "GetNym fails",
			fields: fields{
				methodName: "sov",
				client: &indy.MockIndyClient{
					GetNymErr: errors.New("boom"),
				},
			},
			args:    args{"did:sov:abc123", []vdri.ResolveOpts{}},
			want:    require.Nil,
			wantErr: true,
		},
		{
			name: "Invalid JSON response from GetNym",
			fields: fields{
				methodName: "sov",
				client: &indy.MockIndyClient{
					GetNymValue: &vdr.ReadReply{Data: `_not JSON_`},
					GetNymErr:   nil,
				},
			},
			args:    args{"did:sov:abc123", []vdri.ResolveOpts{}},
			want:    require.Nil,
			wantErr: true,
		},
		{
			name: "Invalid JSON response from GetEndpoint",
			fields: fields{
				methodName: "sov",
				client: &indy.MockIndyClient{
					GetNymValue:    &vdr.ReadReply{Data: `{"dest": "did:sov:abc123", "verkey": "3mJr7AoUCHxNqd"}`},
					GetNymErr:      nil,
					GetEndpointVal: &vdr.ReadReply{Data: `_ not JSON_`},
				},
			},
			args:    args{"did:sov:abc123", []vdri.ResolveOpts{}},
			want:    require.NotNil,
			wantErr: false,
		},
		{
			name: "No endpoint from GetEndpoint",
			fields: fields{
				methodName: "sov",
				client: &indy.MockIndyClient{
					GetNymValue:    &vdr.ReadReply{Data: `{"dest": "did:sov:abc123", "verkey": "3mJr7AoUCHxNqd"}`},
					GetNymErr:      nil,
					GetEndpointVal: &vdr.ReadReply{Data: `{}`},
				},
			},
			args:    args{"did:sov:abc123", []vdri.ResolveOpts{}},
			want:    require.NotNil,
			wantErr: false,
		},
		{
			name: "No nested endpoint from GetEndpoint",
			fields: fields{
				methodName: "sov",
				client: &indy.MockIndyClient{
					GetNymValue:    &vdr.ReadReply{Data: `{"dest": "did:sov:abc123", "verkey": "3mJr7AoUCHxNqd"}`},
					GetNymErr:      nil,
					GetEndpointVal: &vdr.ReadReply{Data: `{"endpoint": {}}`},
				},
			},
			args:    args{"did:sov:abc123", []vdri.ResolveOpts{}},
			want:    require.NotNil,
			wantErr: false,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &VDRI{
				methodName: tt.fields.methodName,
				client:     tt.fields.client,
			}
			got, err := r.Read(tt.args.did, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.want(t, got)
		})
	}

	t.Run("did with no service endpoint", func(t *testing.T) {
		did := "did:sov:abc123"
		indycl := &indy.MockIndyClient{
			GetNymValue:    &vdr.ReadReply{Data: `{"dest": "did:sov:abc123", "verkey": "3mJr7AoUCHxNqd"}`},
			GetNymErr:      nil,
			GetEndpointErr: errors.New("not found"),
		}
		r := &VDRI{
			methodName: "sov",
			client:     indycl,
		}
		doc, err := r.Read(did, vdri.WithNoCache(true))
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Equal(t, did, doc.ID)
		require.NotNil(t, doc.Context)
		require.NotNil(t, doc.Updated)
		require.NotNil(t, doc.Created)
		require.Len(t, doc.Authentication, 1)
		require.Len(t, doc.PublicKey, 1)
		require.Nil(t, doc.Service)

	})

	t.Run("did with service endpoint", func(t *testing.T) {
		did := "did:sov:abc123"
		indycl := &indy.MockIndyClient{
			GetNymValue:    &vdr.ReadReply{Data: `{"dest": "did:sov:abc123", "verkey": "3mJr7AoUCHxNqd"}`},
			GetNymErr:      nil,
			GetEndpointVal: &vdr.ReadReply{Data: `{"endpoint": {"endpoint": "127.0.0.1:8080"}}`},
		}
		r := &VDRI{
			methodName: "sov",
			client:     indycl,
		}
		doc, err := r.Read(did, vdri.WithNoCache(true))
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Equal(t, did, doc.ID)
		require.NotNil(t, doc.Context)
		require.NotNil(t, doc.Updated)
		require.NotNil(t, doc.Created)
		require.NotNil(t, doc.Service)
		require.Len(t, doc.Authentication, 1)
		require.Len(t, doc.PublicKey, 1)

	})

}
