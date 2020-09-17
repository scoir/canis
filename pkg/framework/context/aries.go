/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	mongodbstore "github.com/scoir/canis/pkg/aries/storage/mongodb/store"
	"github.com/scoir/canis/pkg/aries/vdri/indy"
	"github.com/scoir/canis/pkg/credential"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
)

const (
	dbPathKey           = "dbpath"
	wsinboundKey        = "wsinbound"
	defaultMasterKeyURI = "local-lock://default/master/key/"
)

func GetAriesVDRIs(vp *viper.Viper) ([]vdriapi.VDRI, error) {
	var vdri []map[string]interface{}
	err := vp.UnmarshalKey("vdri", &vdri)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get vdri")
	}

	var out []vdriapi.VDRI
	for _, v := range vdri {
		typ, _ := v["type"].(string)
		switch typ {
		case "indy":
			method, _ := v["method"].(string)
			genesisFile, _ := v["genesisFile"].(string)
			re := strings.NewReader(genesisFile)
			indyVDRI, err := indy.New(method, indy.WithIndyVDRGenesisReader(ioutil.NopCloser(re)))
			if err != nil {
				return nil, errors.Wrap(err, "unable to initialize configured indy vdri provider")
			}
			out = append(out, indyVDRI)
		}
	}

	return out, nil
}

type AgentConfig struct {
	framework.Endpoint
	DBPath     string             `mapstructure:"dbpath"`
	WSInbound  framework.Endpoint `mapstructure:"wsinbound"`
	GRPC       framework.Endpoint `mapstructure:"grpc"`
	GRPCBridge framework.Endpoint `mapstructure:"grpcbridge"`
	LedgerURL  string             `mapstructure:"ledgerURL"`

	GetAriesOptions func() []aries.Option

	lock sync.Mutex
}

type kmsProvider struct {
	sp   storage.Provider
	lock secretlock.Service
	kms  kms.KeyManager
}

func (r *kmsProvider) SecretLock() secretlock.Service {
	return r.lock
	// generate a random master key if one does not exist
	// this needs to be in
	//keySize := sha256.Size
	//masterKeyContent := make([]byte, keySize)
	//rand.Read(masterKeyContent)
	//
	//fmt.Println(base64.URLEncoding.EncodeToString(masterKeyContent))
}

func (r *Provider) newProvider() (*kmsProvider, error) {
	out := &kmsProvider{}

	dc := &framework.DatastoreConfig{}
	err := r.vp.UnmarshalKey("datastore", dc)
	if err != nil {
		return nil, errors.Wrap(err, "execution environment is not correctly configured")
	}
	switch dc.Database {
	case "mongo":
		out.sp = mongodbstore.NewProvider(dc.Mongo.URL, dc.Mongo.Database)
	case "postgres":
		//out.sp = pgstore.NewProvider(dc.Postgres.Connection)
	default:
		out.sp = mem.NewProvider()
	}

	mlk := r.vp.GetString("masterLockKey")
	if mlk == "" {
		mlk = "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ="
	}

	out.lock, err = local.NewService(strings.NewReader(mlk), nil)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	out.kms, err = localkms.New(defaultMasterKeyURI, out)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return out, nil
}

func (r *kmsProvider) StorageProvider() storage.Provider {
	return r.sp
}

func (r *kmsProvider) createKMS(_ kms.Provider) (kms.KeyManager, error) {
	return r.kms, nil
}

func (r *Provider) GetAriesContext() *context.Provider {
	if r.ctx == nil {
		err := r.createAriesContext()
		if err != nil {
			log.Fatalln("failed to create aries context", err)
		}

	}
	return r.ctx
}

func (r *Provider) createAriesContext() error {
	fwork, err := aries.New(r.getOptions()...)
	if err != nil {
		return errors.Wrap(err, "failed to start aries agent rest, failed to initialize framework")
	}

	ctx, err := fwork.Context()
	if err != nil {
		return errors.Wrap(err, "failed to start aries agent rest on port, failed to get aries context")
	}
	r.ctx = ctx

	return nil
}

func (r *Provider) getOptions() []aries.Option {
	var out []aries.Option

	p, err := r.newProvider()
	if err != nil {
		panic(err)
	}
	out = append(out, aries.WithStoreProvider(p.StorageProvider()))

	genesisFile := r.vp.GetString("genesisFile")
	if genesisFile != "" {
		genesisData := strings.NewReader(genesisFile)
		indyVDRI, err := indy.New("scr", indy.WithIndyVDRGenesisReader(ioutil.NopCloser(genesisData)))
		if err == nil {
			out = append(out, aries.WithVDRI(indyVDRI))
		}
	}

	if r.vp.IsSet(wsinboundKey) {
		wsinbound := &framework.Endpoint{}
		_ = r.vp.UnmarshalKey(wsinboundKey, wsinbound)
		out = append(out, defaults.WithInboundWSAddr(wsinbound.Address(), wsinbound.Address(), "", ""))
	}

	if r.vp.IsSet("host") && r.vp.IsSet("port") {
		ep := &framework.Endpoint{}
		_ = r.vp.Unmarshal(ep)
		//out = append(out, aries.WithServiceEndpoint(ep.Address()))
	}

	out = append(out, []aries.Option{
		aries.WithKMS(p.createKMS),
		aries.WithMessageServiceProvider(msghandler.NewRegistrar()),
		aries.WithOutboundTransports(ws.NewOutbound()),
	}...)

	return out
}

func (r *Provider) GetDIDClient() (*didexchange.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.didcl != nil {
		return r.didcl, nil
	}

	didcl, err := didexchange.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "error creating did client")
	}

	r.didcl = didcl
	return r.didcl, nil
}

func (r *Provider) GetCredentialClient() (*issuecredential.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.credcl != nil {
		return r.credcl, nil
	}

	credcl, err := issuecredential.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "error creating credential client")
	}
	r.credcl = credcl
	return r.credcl, nil
}

func (r *Provider) GetRouterClient() (*mediator.Client, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.routecl != nil {
		return r.routecl, nil
	}

	routecl, err := mediator.New(r.GetAriesContext())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create route client for college: %v\n")
	}
	r.routecl = routecl
	return r.routecl, nil
}

func (r *Provider) GetSupervisor(h credential.Handler) (*credential.Supervisor, error) {
	sup, err := credential.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create credential supervisor for steward")
	}
	err = sup.Start(h)
	if err != nil {
		return nil, errors.Wrap(err, "unable to start credential supervisor for steward")
	}

	return sup, nil
}

func (r *Provider) GetBouncer() (didex.Bouncer, error) {
	bouncer, err := didex.NewBouncer(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create did supervisor for high school agent")
	}

	return bouncer, nil
}
