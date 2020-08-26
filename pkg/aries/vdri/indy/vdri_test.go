package indy

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scoir/canis/pkg/mock/vdri/indy"
)

func TestNew(t *testing.T) {
	type args struct {
		methodName string
		opts       []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *VDRI
		wantErr bool
	}{
		{
			name: "with indy client and refresh",
			args: args{
				methodName: "sov",
				opts:       []Option{WithIndyClient(&indy.MockIndyClient{}), WithRefresh(true)},
			},
			want: &VDRI{
				methodName: "sov",
				refresh:    true,
				client:     &indy.MockIndyClient{},
			},
			wantErr: false,
		},
		{
			name: "pool refresh fails",
			args: args{
				methodName: "sov",
				opts: []Option{WithIndyClient(&indy.MockIndyClient{
					RefreshErr: errors.New("boom"),
				})},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "with no indy client",
			args: args{
				methodName: "sov",
				opts:       []Option{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.methodName, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVDRI_Accept(t *testing.T) {
	type fields struct {
		methodName string
	}
	type args struct {
		method string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "matching method",
			fields: fields{
				methodName: "sov",
			},
			args: args{
				method: "sov",
			},
			want: true,
		},
		{
			name: "mismatching method",
			fields: fields{
				methodName: "sov",
			},
			args: args{
				method: "ioe",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &VDRI{
				methodName: tt.fields.methodName,
			}
			if got := r.Accept(tt.args.method); got != tt.want {
				t.Errorf("Accept() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVDRI_Close(t *testing.T) {
	type fields struct {
		client indyClient
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Close works",
			fields: fields{
				client: &indy.MockIndyClient{
					CloseErr: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Close fails",
			fields: fields{
				client: &indy.MockIndyClient{
					CloseErr: errors.New("boom"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &VDRI{
				client: tt.fields.client,
			}
			if err := r.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVDRI_Store(t *testing.T) {
	t.Run("no op", func(t *testing.T) {
		r := &VDRI{}

		err := r.Store(nil, nil)
		require.NoError(t, err)
	})
}

func TestWithRefresh(t *testing.T) {
	t.Run("refresh sets the value", func(t *testing.T) {
		refresh := WithRefresh(true)
		opts := &VDRI{}
		refresh(opts)
		require.True(t, opts.refresh)
	})
}
