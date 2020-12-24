/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hyperledger/aries-framework-go-ext/component/didcomm/transport/amqp"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/dispatcher"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/presentproof"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/config"
	credindyengine "github.com/scoir/canis/pkg/credential/engine/indy"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
	"github.com/scoir/canis/pkg/presentproof/engine"
	"github.com/scoir/canis/pkg/presentproof/engine/indy"
	"github.com/scoir/canis/pkg/ursa"
)

var (
	cfgFile        string
	ctx            *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm verifier.",
	Long: `"The canis didcomm verifier but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	lock                 secretlock.Service
	store                datastore.Store
	ariesStorageProvider storage.Provider
	keyMgr               kms.KeyManager
	conf                 config.Config
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	configProvider = &config.ViperConfigProvider{
		DefaultConfigName: "canis-verifier-config",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-verifier-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	conf := configProvider.Load(cfgFile).
		WithDatastore().
		WithLedgerStore().
		WithMasterLockKey().
		WithAMQP().
		WithVDRI().
		WithIndyRegistry()

	dc, err := conf.DataStore()
	if err != nil {
		log.Fatalln("invalid datastore key in configuration", err)
	}

	sp, err := dc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := sp.Open()
	if err != nil {
		log.Fatalln("unable to open datastore")
	}

	lc, err := conf.LedgerStore()
	if err != nil {
		log.Fatalln("invalid ledgerstore key in configuration")
	}

	ls, err := lc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	lock, err := local.NewService(strings.NewReader(conf.MasterLockKey()), nil)
	if err != nil {
		log.Fatalln("error creating lock service")
	}

	ctx = &Provider{
		lock:                 lock,
		store:                store,
		ariesStorageProvider: ls,
		conf:                 conf,
	}

	ctx.keyMgr, err = localkms.New("local-lock://default/master/key/", ctx)
	if err != nil {
		log.Fatalln("unable to create local kms")
	}
}

func (r *Provider) Store() datastore.Store {
	return r.store
}

// GetStorageProvider todo
func (r *Provider) StorageProvider() storage.Provider {
	return r.ariesStorageProvider
}

// GetGRPCEndpoint todo
func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("api.grpc")
}

// GetBridgeEndpoint todo
func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("api.grpcBridge")
}

// GetAriesContext todo
func (r *Provider) GetAriesContext() (*ariescontext.Provider, error) {
	external := r.conf.GetString("inbound.external")
	cfg, err := r.conf.AMQPConfig()
	if err != nil {
		return nil, err
	}

	vdrisConfig, err := r.conf.VDRIs()
	if err != nil {
		return nil, err
	}

	vdris, err := context.GetAriesVDRIs(vdrisConfig)
	if err != nil {
		return nil, err
	}

	amqpInbound, err := amqp.NewInbound(cfg.Endpoint(), external, "present-proof", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "amqp.NewInbound")
	}

	vopts := []aries.Option{
		aries.WithStoreProvider(r.ariesStorageProvider),
		aries.WithInboundTransport(amqpInbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithSecretLock(r.lock),
		aries.WithProtocols(newPresentProofSvc()),
	}
	for _, vdri := range vdris {
		vopts = append(vopts, aries.WithVDRI(vdri))
	}

	ar, err := aries.New(vopts...)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create aries defaults")
	}

	actx, err := ar.Context()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries context")
	}

	return actx, err
}

func newPresentProofSvc() api.ProtocolSvcCreator {
	return func(prv api.Provider) (dispatcher.ProtocolService, error) {
		svc, err := presentproof.New(prv)
		if err != nil {
			return nil, err
		}

		// sets default middleware to the service
		// svc.Use(mdissuecredential.SaveCredentials(prv))

		return svc, nil
	}
}

// IndyVDR todo
func (r *Provider) IndyVDR() (credindyengine.VDRClient, error) {
	genesisFile := r.conf.GetString("registry.indy.genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	return cl, nil
}

// KMS todo
func (r *Provider) KMS() kms.KeyManager {
	return r.keyMgr
}

// GetPresentationEngineRegistry todo
func (r *Provider) GetPresentationEngineRegistry() (engine.PresentationRegistry, error) {
	e, err := indy.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get presentation engine registry")
	}
	return engine.New(r, engine.WithEngine(e)), nil
}

// SecretLock todo
func (r *Provider) SecretLock() secretlock.Service {
	return r.lock
}

func (r *Provider) Oracle() credindyengine.Oracle {
	return &ursa.CryptoOracle{}
}
