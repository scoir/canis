package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/client/mediator"
	arieslog "github.com/hyperledger/aries-framework-go/pkg/common/log"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/transport/ws"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/defaults"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"
	mongodbstore "github.com/scoir/aries-storage-mongo/pkg"

	vindy "github.com/scoir/canis/pkg/aries/vdri/indy"
	didex "github.com/scoir/canis/pkg/didexchange"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/protogen/common"
)

var ctx *ariescontext.Provider
var bouncer didex.Bouncer
var mediatorConnection *didexchange.Connection
var vdrclient *vdr.Client
var edgeAgentStore storage.Store

func main() {
	arieslog.SetLevel("aries-framework/out-of-band/service", arieslog.CRITICAL)
	arieslog.SetLevel("aries-framework/ws", arieslog.CRITICAL)
	//arieslog.SetLevel("aries-framework/did-exchange/service", arieslog.DEBUG)
	//arieslog.SetLevel("aries-framework/issuecredential/service", arieslog.DEBUG)
	arieslog.SetLevel("aries-framework/presentproof/service", arieslog.DEBUG)

	createAriesContext()

	log.Println("checking mediator connection")
	if mediatorConnection == nil {
		log.Println("creating mediator connection")
		connect()
		log.Println("mediator connection complete")
	}

	log.Println("establishing routing")
	establishRouting()
	log.Println("routing established")

	didcli, err := didexchange.New(ctx)
	if err != nil {
		log.Fatalln("unable to create did client", err)
	}

	invite, err := didcli.CreateInvitation("testing")
	if err != nil {
		log.Fatalln("unable to create invitation")
	}

	d, _ := json.MarshalIndent(invite, " ", " ")
	fmt.Println(string(d))

}

func establishRouting() {
	rcli, err := mediator.New(ctx)
	if err != nil {
		log.Fatalln("unable to create mediator client", err)
	}

	err = rcli.Register(mediatorConnection.ConnectionID)
	if err != nil {
		log.Fatalln("unable to register for routing with mediator", err)
	}

	log.Println("Successful routing config")
}

type registration struct {
	DID        string `json:"did"`
	ExternalID string `json:"external_id"`
	Secret     string `json:"secret"`
}

func connect() {

	reg := &registration{
		ExternalID: "test-edge-agent",
		Secret:     "ArwXoACJgOleVZ2PY7kXn7rA0II0mHYDhc6WrBH8fDAc",
	}

	w := &bytes.Buffer{}
	enc := json.NewEncoder(w)
	_ = enc.Encode(reg)
	req, err := http.NewRequest("POST", "http://local.scoir.com:7779/edge/agents/register", w)
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}
	req.Header.Set("X-API-Key", "D3YYdahdgC7VZeJwP4rhZcozCRHsqQT3VKxK9hTc2Yoh")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Println(string(b))

	inviteResponse := &common.InvitationResponse{}
	err = json.Unmarshal(b, inviteResponse)
	if err != nil {
		log.Fatalln("bad edge agent register respose", err)
	}

	fmt.Println(inviteResponse.Invitation)

	invite := &didexchange.Invitation{}
	err = json.NewDecoder(strings.NewReader(inviteResponse.Invitation)).Decode(invite)
	if err != nil {
		log.Fatalln("invalid invitation", err)
	}

	mediatorConnection, err = bouncer.EstablishConnection(invite, 10*time.Second)
	if err != nil {
		log.Fatalln("Error requesting invitation from mediator", err)
	}

	d, _ := json.MarshalIndent(mediatorConnection, " ", " ")
	_ = edgeAgentStore.Put("mediator", d)

}

func createAriesContext() {
	wsinbound := "172.17.0.1:3001"

	genesis, err := os.Open("./deploy/canis-chart/indy/genesis.txn")
	if err != nil {
		log.Fatalln("unable to open genesis file", err)
	}
	vdrclient, err = vdr.New(genesis)
	if err != nil {
		log.Fatalln("unable to connect to indy vdr", err)
	}

	storeProv := mongodbstore.NewProvider("mongodb://172.17.0.1:27017", mongodbstore.WithDBPrefix("subject"))
	edgeAgentStore, _ = storeProv.OpenStore("connections")
	indyVDRI, err := vindy.New("sov", vindy.WithIndyClient(vdrclient))
	if err != nil {
		log.Fatalln("unable to create aries indy vdr", err)
	}

	ar, err := aries.New(
		aries.WithStoreProvider(storeProv),
		defaults.WithInboundWSAddr(wsinbound, fmt.Sprintf("ws://%s", wsinbound), "", ""),
		aries.WithOutboundTransports(ws.NewOutbound()),
		aries.WithVDRI(indyVDRI),
	)
	if err != nil {
		log.Fatalln("Unable to create", err)
	}

	ctx, err = ar.Context()
	if err != nil {
		log.Fatalln("unable to get context", err)
	}

	prov := framework.NewSimpleProvider(ctx)
	bouncer, err = didex.NewBouncer(prov)
	if err != nil {
		log.Fatalln("could not create bouncer", err)
	}

	d, err := edgeAgentStore.Get("mediator")
	if err == nil {
		mediatorConnection = &didexchange.Connection{}
		err = json.Unmarshal(d, mediatorConnection)
		if err != nil {
			log.Fatalln("issuer conneciton stored but not valid")
		}
	}

}
