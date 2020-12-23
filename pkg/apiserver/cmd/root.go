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
	"google.golang.org/grpc"

	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	"github.com/scoir/canis/pkg/config"
	cengine "github.com/scoir/canis/pkg/credential/engine"
	credengine "github.com/scoir/canis/pkg/credential/engine"
	credindyengine "github.com/scoir/canis/pkg/credential/engine/indy"
	credldsengine "github.com/scoir/canis/pkg/credential/engine/lds"
	"github.com/scoir/canis/pkg/datastore"
	doormanapi "github.com/scoir/canis/pkg/didcomm/doorman/api/protogen"
	issuerapi "github.com/scoir/canis/pkg/didcomm/issuer/api/protogen"
	lbapi "github.com/scoir/canis/pkg/didcomm/loadbalancer/api/protogen"
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
	cfgFile        string
	ctx            *Provider
	configProvider config.Provider
)

var rootCmd = &cobra.Command{
	Use:   "canis-apiserver",
	Short: "The canis steward orchestration service.",
	Long: `"The canis steward orchestration service.".

 Find more information at: https://canis.io/docs/reference/canis/overview`,
}

type Provider struct {
	lock    secretlock.Service
	store   datastore.Store
	ariesSP storage.Provider
	vdriReg vdriapi.Registry
	keyMgr  kms.KeyManager
	conf    config.Config
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
		DefaultConfigName: "canis-apiserver-config",
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/canis/canis-apiserver-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	conf := configProvider.
		Load(cfgFile).
		WithDatastore().
		WithLedgerStore().
		WithMasterLockKey().
		WithVDRI().
		WithLedgerGenesis()

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
		lock:    lock,
		store:   store,
		ariesSP: ls,
		conf:    conf,
	}

	ctx.keyMgr, err = localkms.New("local-lock://default/master/key/", ctx)
	if err != nil {
		log.Fatalln("unable to create local kms", err)
	}

	vdrisConfig, err := conf.VDRIs()
	if err != nil {
		log.Fatalln("unable to load aries config", err)
	}

	vdris, err := context.GetAriesVDRIs(vdrisConfig)
	if err != nil {
		log.Fatalln("unable to get aries vdris", err)
	}

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

func (r *Provider) Issuer() credindyengine.UrsaIssuer {
	return credindyengine.NewIssuer(r)
}

func (r *Provider) Oracle() ursa.Oracle {
	return &ursa.CryptoOracle{}
}

func (r *Provider) IndyVDR() (indywrapper.IndyVDRClient, error) {
	genesisFile := r.conf.LedgerGenesis()
	re := strings.NewReader(genesisFile)
	cl, err := vdr.New(ioutil.NopCloser(re))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get indy vdr client")
	}

	return cl, nil
}

func (r *Provider) GetGRPCEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("api.grpc")
}

func (r *Provider) GetBridgeEndpoint() (*framework.Endpoint, error) {
	return r.conf.Endpoint("api.grpcBridge")
}

func (r *Provider) GetDoormanClient() (doormanapi.DoormanClient, error) {
	ep, err := r.conf.Endpoint("doorman.grpc")
	if err != nil {
		return nil, errors.Wrap(err, "doorman grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for doorman client")
	}
	cl := doormanapi.NewDoormanClient(cc)
	return cl, nil
}

func (r *Provider) GetIssuerClient() (issuerapi.IssuerClient, error) {
	ep, err := r.conf.Endpoint("issuer.grpc")
	if err != nil {
		return nil, errors.Wrap(err, "issuer grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for issuer client")
	}
	cl := issuerapi.NewIssuerClient(cc)
	return cl, nil
}

func (r *Provider) GetVerifierClient() (verifier.VerifierClient, error) {
	ep, err := r.conf.Endpoint("verifier.grpc")
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

func (r *Provider) GetLoadbalancerClient() (lbapi.LoadbalancerClient, error) {
	ep, err := r.conf.Endpoint("loadbalancer.grpc")
	if err != nil {
		return nil, errors.Wrap(err, "loadbalancer grpc is not properly configured")
	}

	cc, err := grpc.Dial(ep.Address(), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc for load balancer client")
	}
	lb := lbapi.NewLoadbalancerClient(cc)
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

func (r *Provider) Verifier() presentindyengine.Verifier {
	return presentindyengine.NewVerifier(r.Store())
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
