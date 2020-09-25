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

	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/scoir/canis/pkg/aries/transport/amqp"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
	indywrapper "github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
	"github.com/scoir/canis/pkg/presentproof/engine"
	"github.com/scoir/canis/pkg/presentproof/engine/indy"
	"github.com/scoir/canis/pkg/ursa"
)

var (
	cfgFile string
	ctx     *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-didcomm",
	Short: "The canis didcomm verifier.",
	Long: `"The canis didcomm verifier but longer.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp                   *viper.Viper
	lock                 secretlock.Service
	store                datastore.Store
	ariesStorageProvider storage.Provider
}

func (r *Provider) Issuer() ursa.Issuer {
	panic("implement me")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-didcomm-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	vp := viper.New()
	if cfgFile != "" {
		// Use vp file from the flag.
		vp.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		vp.SetConfigType("yaml")
		vp.AddConfigPath("/etc/canis/")
		vp.AddConfigPath("./deploy/compose/")
		vp.SetConfigName("canis-didcomm-config")
	}

	vp.SetEnvPrefix("CANIS")
	vp.AutomaticEnv() // read in environment variables that match
	_ = vp.BindPFlags(pflag.CommandLine)

	// If a vp file is found, read it in.
	if err := vp.ReadInConfig(); err != nil {
		fmt.Println("unable to read vp:", vp.ConfigFileUsed(), err)
		os.Exit(1)
	}

	dc := &framework.DatastoreConfig{}
	err := vp.UnmarshalKey("datastore", dc)
	if err != nil {
		log.Fatalln("invalid datastore key in configuration")
	}

	sp, err := dc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := sp.Open()
	if err != nil {
		log.Fatalln("unable to open datastore")
	}

	lc := &framework.LedgerStoreConfig{}
	err = vp.UnmarshalKey("ledgerstore", lc)
	if err != nil {
		log.Fatalln("invalid ledgerstore key in configuration")
	}

	ls, err := lc.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}
	mlk := vp.GetString("masterLockKey")
	if mlk == "" {
		mlk = "OTsonzgWMNAqR24bgGcZVHVBB_oqLoXntW4s_vCs6uQ="
	}

	lock, err := local.NewService(strings.NewReader(mlk), nil)
	if err != nil {
		log.Fatalln("error creating lock service")
	}

	ctx = &Provider{
		vp:                   vp,
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
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc is not properly configured")
	}

	return ep, nil
}

// GetBridgeEndpoint todo
func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpcBridge", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc bridge is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetAriesContext() (*ariescontext.Provider, error) {
	external := r.vp.GetString("inbound.external")
	config := &framework.AMQPConfig{}
	err := r.vp.UnmarshalKey("inbound.amqp", config)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal inbound.amqp")
	}

	ariesSub := r.vp.Sub("aries")
	vdris, err := context.GetAriesVDRIs(ariesSub)
	if err != nil {
		return nil, errors.Wrap(err, "GetAriesVDRIs")
	}

	amqpInbound, err := amqp.NewInbound(config.Endpoint(), external, "present-proof", "", "")
	if err != nil {
		return nil, errors.Wrap(err, "amqp.NewInbound")
	}

	vopts := []aries.Option{
		aries.WithStoreProvider(r.ariesStorageProvider),
		aries.WithInboundTransport(amqpInbound),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithSecretLock(r.lock),
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

func (r *Provider) Verifier() ursa.Verifier {
	return ursa.NewVerifier()
}

func (r *Provider) IndyVDR() (indywrapper.IndyVDRClient, error) {
	genesisFile := r.vp.GetString("presentproof.indy.genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	return cl, nil
}

func (r *Provider) KMS() (kms.KeyManager, error) {
	mgr, err := localkms.New("local-lock://default/master/key/", r)
	return mgr, errors.Wrap(err, "unable to create locakkms")
}

func (r *Provider) GetPresentationEngineRegistry() (engine.PresentationRegistry, error) {
	e, err := indy.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get presentation engine registry")
	}
	return engine.New(r, engine.WithEngine(e)), nil
}

func (r *Provider) SecretLock() secretlock.Service {
	return r.lock
}