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

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/messaging/msghandler"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/storage/mem"
	"github.com/pkg/errors"
	mongodbstore "github.com/scoir/aries-storage-mongo/pkg"

	"github.com/scoir/canis/pkg/aries/vdri/indy"
	"github.com/scoir/canis/pkg/credential"
)

const (
	defaultMasterKeyURI = "local-lock://default/master/key/"
)

func GetAriesVDRIs(vdriConfig []map[string]interface{}) ([]vdriapi.VDRI, error) {

	var out []vdriapi.VDRI
	for _, v := range vdriConfig {
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

	cfg := r.conf.WithDatastore().
		WithMasterLockKey().
		WithVDRI()

	dc, err := cfg.DataStore()
	if err != nil {
		return nil, err
	}

	switch dc.Database {
	case "mongo":
		out.sp = mongodbstore.NewProvider(dc.Mongo.URL)
	default:
		out.sp = mem.NewProvider()
	}

	out.lock, err = local.NewService(strings.NewReader(r.conf.MasterLockKey()), nil)
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

	vdriConfig, err := r.conf.VDRIs()
	if err != nil {
		log.Println("failed to load VDRI config", err)
	}

	if err == nil {
		for _, v := range vdriConfig {
			typ, _ := v["type"].(string)
			switch typ {
			case "indy":
				method, _ := v["method"].(string)
				genesisFile, _ := v["genesisFile"].(string)
				re := strings.NewReader(genesisFile)
				indyVDRI, err := indy.New(method, indy.WithIndyVDRGenesisReader(ioutil.NopCloser(re)))
				if err == nil {
					out = append(out, aries.WithVDRI(indyVDRI))
				}

			}
		}
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
