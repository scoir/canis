package did

import (
	"crypto/ed25519"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMyDid(t *testing.T) {
	type args struct {
		info *MyDIDInfo
	}
	tests := []struct {
		name       string
		args       args
		wantDID    string
		wantVerkey string
		wantErr    bool
	}{
		{
			name: "test seed, short",
			args: args{
				info: &MyDIDInfo{
					Seed:       "b2352b32947e188eb72871093ac6217e",
					Cid:        true,
					MethodName: "sov",
				},
			},
			wantDID:    "did:sov:WvRwKqxFLtJ3YbhmHZBpmy",
			wantVerkey: "HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantErr:    false,
		},
		{
			name: "test seed, short, no method",
			args: args{
				info: &MyDIDInfo{
					Seed: "b2352b32947e188eb72871093ac6217e",
					Cid:  true,
				},
			},
			wantDID:    "did:WvRwKqxFLtJ3YbhmHZBpmy",
			wantVerkey: "HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantErr:    false,
		},
		{
			name: "test seed, existing DID",
			args: args{
				info: &MyDIDInfo{
					DID:        "abc123",
					MethodName: "scr",
					Seed:       "b2352b32947e188eb72871093ac6217e",
					Cid:        true,
				},
			},
			wantDID:    "did:scr:abc123",
			wantVerkey: "HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantErr:    false,
		},
		{
			name: "test seed, long",
			args: args{
				info: &MyDIDInfo{
					Seed:       "b2352b32947e188eb72871093ac6217e",
					Cid:        false,
					MethodName: "sov",
				},
			},
			wantDID:    "did:sov:HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantVerkey: "HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantErr:    false,
		},
		{
			name: "test no seed",
			args: args{
				info: &MyDIDInfo{
					Seed:       "",
					Cid:        true,
					MethodName: "sov",
				},
			},
			wantErr: false,
		},
		{
			name: "test bad base64 seed",
			args: args{
				info: &MyDIDInfo{
					Seed: "cG9vcA==",
				},
			},
			wantErr: true,
		},
		{
			name: "test base64 seed",
			args: args{
				info: &MyDIDInfo{
					Seed:       "YjIzNTJiMzI5NDdlMTg4ZWI3Mjg3MTA5M2FjNjIxN2U=",
					Cid:        false,
					MethodName: "ioe",
				},
			},
			wantDID:    "did:ioe:HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantVerkey: "HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantErr:    false,
		},
		{
			name: "test hex seed",
			args: args{
				info: &MyDIDInfo{
					Seed:       "6232333532623332393437653138386562373238373130393361633632313765",
					Cid:        true,
					MethodName: "hex",
				},
			},
			wantDID:    "did:hex:WvRwKqxFLtJ3YbhmHZBpmy",
			wantVerkey: "HJsMyfABm7gmPse8QzgUePRwTbQRyALgeZudJuYbYmro",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			did, _, err := CreateMyDid(tt.args.info)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			if tt.wantDID != "" {
				require.Equal(t, tt.wantDID, did.String())
			}

			if tt.wantVerkey != "" {
				require.Equal(t, tt.wantVerkey, did.Verkey)
			}
		})
	}
}

func TestKeyPair(t *testing.T) {
	type fields struct {
		vk string
		sk string
	}
	tests := []struct {
		name       string
		fields     fields
		want       ed25519.PrivateKey
		wantVerkey string
	}{
		{
			name: "values",
			fields: fields{
				vk: "test",
				sk: "BzfBWFc",
			},
			want:       []byte("aries"),
			wantVerkey: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &KeyPair{
				vk: tt.fields.vk,
				sk: tt.fields.sk,
			}
			if got := r.Priv(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Priv() = %v, want %v", string(got), tt.want)
			}
			if got := r.Verkey(); !reflect.DeepEqual(got, tt.wantVerkey) {
				t.Errorf("Verkey() = %v, want %v", got, tt.wantVerkey)
			}
		})
	}
}
