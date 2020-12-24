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
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/aries/didcomm/protocol/middleware/issuecredential"
	"github.com/scoir/canis/pkg/config"
	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/credential/engine/indy"
	"github.com/scoir/canis/pkg/credential/engine/lds"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
	"github.com/scoir/canis/pkg/ursa"
)

var (
	cfgFile        string
	ctx            *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm issuer service.",
	Long: `"The canis didcomm issuer service but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	actx                 *ariescontext.Provider
	lock                 secretlock.Service
	store                datastore.Store
	ariesStorageProvider storage.Provider
	conf                 config.Config
}

func (r *Provider) Oracle() indy.Oracle {
	return &ursa.CryptoOracle{}
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
		DefaultConfigName: "canis-issuer-config",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-issuer-config.yaml)")
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
		conf:                 conf,
		lock:                 lock,
		store:                store,
		ariesStorageProvider: ls,
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

func (r *Provider) VDRIRegistry() vdriapi.Registry {
	ctx, err := r.GetAriesContext()
	if err != nil {
		log.Fatalln("unable to load aries context")
	}

	return ctx.VDRIRegistry()
}

func (r *Provider) GetAriesContext() (*ariescontext.Provider, error) {
	if r.actx != nil {
		return r.actx, nil
	}

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

	amqpInbound, err := amqp.NewInbound(cfg.Endpoint(), external, "issue-credential", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create amqp aries inbound")
	}

	vopts := []aries.Option{
		aries.WithStoreProvider(r.ariesStorageProvider),
		aries.WithInboundTransport(amqpInbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithSecretLock(r.lock),
		aries.WithProtocols(r.newIssueCredentialSvc()),
	}
	for _, vdri := range vdris {
		vopts = append(vopts, aries.WithVDRI(vdri))
	}

	ar, err := aries.New(vopts...)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create aries defaults")
	}

	r.actx, err = ar.Context()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get aries context")
	}

	return r.actx, err
}

func (r *Provider) newIssueCredentialSvc() api.ProtocolSvcCreator {
	return func(prv api.Provider) (dispatcher.ProtocolService, error) {
		service, err := icprotocol.New(prv)
		if err != nil {
			return nil, err
		}

		// sets default middleware to the service
		service.Use(issuecredential.SaveCredentials(r))

		return service, nil
	}
}

func (r *Provider) IndyVDR() (indy.VDRClient, error) {
	genesisFile := r.conf.GetString("registry.indy.genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	return cl, nil
}

func (r *Provider) KMS() kms.KeyManager {
	mgr, err := localkms.New("local-lock://default/master/key/", r)
	if err != nil {
		log.Fatalln("unable to create local kms")
	}
	return mgr
}

func (r *Provider) GetCredentialEngineRegistry() (engine.CredentialRegistry, error) {
	e, err := indy.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy credential engine")
	}

	ldse, err := lds.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get LDS credential engine")
	}
	return engine.New(r, engine.WithEngine(e), engine.WithEngine(ldse)), nil
}

func (r *Provider) SecretLock() secretlock.Service {
	return r.lock
}
