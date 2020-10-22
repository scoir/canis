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

	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/local"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/aries-framework-go/pkg/vdri"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	cengine "github.com/scoir/canis/pkg/credential/engine"
	credengine "github.com/scoir/canis/pkg/credential/engine"
	credindyengine "github.com/scoir/canis/pkg/credential/engine/indy"
	credldsengine "github.com/scoir/canis/pkg/credential/engine/lds"
	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/didcomm/doorman/api"
	issuer "github.com/scoir/canis/pkg/didcomm/issuer/api"
	loadbalancer "github.com/scoir/canis/pkg/didcomm/loadbalancer/api"
	verifier "github.com/scoir/canis/pkg/didcomm/verifier/api/protogen"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/framework/context"
	indywrapper "github.com/scoir/canis/pkg/indy"
	presentengine "github.com/scoir/canis/pkg/presentproof/engine"
	presentindyengine "github.com/scoir/canis/pkg/presentproof/engine/indy"
	presentjsonldengine "github.com/scoir/canis/pkg/presentproof/engine/jsonld"
	"github.com/scoir/canis/pkg/ursa"
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
	vp      *viper.Viper
	lock    secretlock.Service
	store   datastore.Store
	ariesSP storage.Provider
	vdriReg vdriapi.Registry
	keyMgr  kms.KeyManager
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
		vp:      vp,
		lock:    lock,
		store:   store,
		ariesSP: ls,
	}

	ctx.keyMgr, err = localkms.New("local-lock://default/master/key/", ctx)
	if err != nil {
		log.Fatalln("unable to create local kms")
	}

	ariesSub := vp.Sub("aries")
	vdris, err := context.GetAriesVDRIs(ariesSub)
	var vopts []vdri.Option
	for _, v := range vdris {
		vopts = append(vopts, vdri.WithVDRI(v))
	}

	ctx.vdriReg = vdri.New(ctx, vopts...)

}

func (r *Provider) StorageProvider() storage.Provider {
	return r.ariesSP
}

func (r *Provider) Store() datastore.Store {
	return r.store
}

func (r *Provider) Issuer() ursa.Issuer {
	return ursa.NewIssuer()
}

func (r *Provider) Verifier() ursa.Verifier {
	return ursa.NewVerifier()
}

func (r *Provider) IndyVDR() (indywrapper.IndyVDRClient, error) {
	genesisFile := r.vp.GetString("genesisFile")
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

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
	if err != nil {
		return nil, errors.Wrap(err, "doorman grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for doorman client")
	}
	cl := api.NewDoormanClient(cc)
	return cl, nil
}

func (r *Provider) GetIssuerClient() (issuer.IssuerClient, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("issuer.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "issuer grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for issuer client")
	}
	cl := issuer.NewIssuerClient(cc)
	return cl, nil
}

func (r *Provider) GetVerifierClient() (verifier.VerifierClient, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("verifier.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "verifier grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for verifier client")
	}
	vc := verifier.NewVerifierClient(cc)
	return vc, nil
}

func (r *Provider) GetLoadbalancerClient() (loadbalancer.LoadbalancerClient, error) {
	ep := &framework.Endpoint{}
	err := r.vp.UnmarshalKey("loadbalancer.grpc", ep)
	if err != nil {
		return nil, errors.Wrap(err, "loadbalancer grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for load balancer client")
	}
	lb := loadbalancer.NewLoadbalancerClient(cc)
	return lb, nil
}

func (r *Provider) KMS() kms.KeyManager {
	return r.keyMgr
}

func (r *Provider) GetCredentialEngineRegistry() (credengine.CredentialRegistry, error) {
	cie, err := credindyengine.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy credential engine")
	}

	ldse, err := credldsengine.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get LDS credential engine")
	}

	return credengine.New(r, credengine.WithEngine(cie), cengine.WithEngine(ldse)), nil
}

func (r *Provider) VDRIRegistry() vdriapi.Registry {
	return r.vdriReg
}

func (r *Provider) GetPresentationEngineRegistry() (presentengine.PresentationRegistry, error) {
	pie, err := presentindyengine.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create indy presentation engine")
	}

	pjlde, err := presentjsonldengine.New(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create json-ld presentation engine")
	}

	return presentengine.New(r, presentengine.WithEngine(pie), presentengine.WithEngine(pjlde)), nil
}

func (r *Provider) SecretLock() secretlock.Service {
	return r.lock
}
