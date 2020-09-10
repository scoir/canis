/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/scoir/canis/pkg/credential/engine"
	"github.com/scoir/canis/pkg/credential/engine/indy"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/datastore/manager"
	"github.com/scoir/canis/pkg/didcomm/doorman/api"
	issuer "github.com/scoir/canis/pkg/didcomm/issuer/api"
	loadbalancer "github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

var (
	cfgFile string
	ctx     *Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-apiserver",
	Short: "The canis steward orchestration service.",
	Long: `"The canis steward orchestration service.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	vp                   *viper.Viper
	lock                 secretlock.Service
	store                datastore.Store
	ariesStorageProvider storage.Provider
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-apiserver-config.yaml)")
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
		vp.AddConfigPath("./config/docker/")
		vp.SetConfigName("canis-apiserver-config")
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

	dm := manager.NewDataProviderManager(dc)
	sp, err := dm.DefaultStoreProvider()
	if err != nil {
		log.Fatalln("unable to get default data store", err)
	}

	store, err := sp.OpenStore("canis")
	if err != nil {
		log.Fatalln("unable to get default data store", err)
	}

	asp, err := store.GetAriesProvider()
	if err != nil {
		log.Fatalln("unable to load aries storage provider", err)
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
		ariesStorageProvider: asp,
	}
}

func (r *Provider) StorageProvider() storage.Provider {
	return r.ariesStorageProvider
}

func (r *Provider) Store() datastore.Store {
	return r.store
}

func (r *Provider) IndyVDR() (*vdr.Client, error) {
	genesisFile := r.vp.GetString("genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	status, err := cl.GetPoolStatus()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get pool status")
	}

	d, _ := json.MarshalIndent(status, " ", " ")
	fmt.Println(string(d))

	return cl, nil
}

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("api.grpcBridge", ep)
	if err != nil {
		return nil, errors.Wrap(err, "grpc bridge is not properly configured")
	}

	return ep, nil
}

func (r *Provider) GetDoormanClient() (api.DoormanClient, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("doorman.grpc", ep)

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for steward client")
	}
	cl := api.NewDoormanClient(cc)
	return cl, nil
}

func (r *Provider) GetIssuerClient() (issuer.IssuerClient, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("issuer.grpc", ep)

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for steward client")
	}
	cl := issuer.NewIssuerClient(cc)
	return cl, nil
}

func (r *Provider) GetLoadbalancerClient() (loadbalancer.LoadbalancerClient, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("loadbalancer.grpc", ep)

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for steward client")
	}
	lb := loadbalancer.NewLoadbalancerClient(cc)
	return lb, nil
}

func (r *Provider) KMS() (kms.KeyManager, error) {
	mgr, err := localkms.New("", r)
	return mgr, errors.Wrap(err, "unable to create locakkms")
}

func (r *Provider) GetCredentailEngineRegistry() (engine.CredentialRegistry, error) {
	e, err := indy.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get credential engine registry")
	}
	return engine.New(r, engine.WithEngine(e)), nil
}

func (r *Provider) SecretLock() secretlock.Service {
	return r.lock
}
