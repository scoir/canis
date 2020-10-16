package resolver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/btcsuite/btcutil/base58"
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	"github.com/pkg/errors"
	goji "goji.io"
	"goji.io/pat"

	indywrapper "github.com/scoir/canis/pkg/indy"
	"github.com/scoir/canis/pkg/util"
)

const (
	context  = "https://www.w3.org/ns/did-resolution/v1"
	schemaV1 = "https://w3id.org/did/v1"
	keyType  = "Ed25519VerificationKey2018"
	driverID = "did:%s"
	driver   = "CanisHttpIndyDriver"
)

type HTTPIndyResolver struct {
	addr       string
	methodName string
	client     indywrapper.IndyVDRClient
}

type provider interface {
	IndyVDR() indywrapper.IndyVDRClient
}

type didResolution struct {
	Context          interface{}            `json:"@context"`
	DIDDocument      map[string]interface{} `json:"didDocument"`
	ResolverMetadata map[string]interface{} `json:"resolverMetadata"`
	MethodMetadata   map[string]interface{} `json:"methodMetadata"`
}

func NewHTTPIndyResolver(addr, method string, ctx provider) *HTTPIndyResolver {
	r := &HTTPIndyResolver{
		addr:       addr,
		methodName: method,
		client:     ctx.IndyVDR(),
	}

	return r
}

func (r *HTTPIndyResolver) Start() error {
	mux := goji.NewMux()
	mux.Handle(pat.Get("/did/:did"), http.HandlerFunc(r.resolve))

	log.Println("http indy resolver listening on", r.addr)
	return http.ListenAndServe(r.addr, mux)
}

func (r *HTTPIndyResolver) resolve(w http.ResponseWriter, req *http.Request) {
	did := pat.Param(req, "did")

	out, err := r.Read(did)
	if err != nil {
		util.WriteError(w, err.Error())
		return
	}

	d, _ := json.MarshalIndent(out, " ", " ")
	util.WriteSuccess(w, d)
}

func (r *HTTPIndyResolver) Read(did string) (*didResolution, error) {
	start := time.Now()
	parsedDID, err := diddoc.Parse(did)
	if err != nil {
		return nil, fmt.Errorf("parsing did failed in indy resolver: (%w)", err)
	}

	if parsedDID.Method != r.methodName {
		return nil, fmt.Errorf("invalid indy method ID: %s", parsedDID.MethodSpecificID)
	}

	rply, err := r.client.GetNym(parsedDID.MethodSpecificID)
	if err != nil {
		return nil, err
	}

	nymResp := map[string]interface{}{}
	err = json.Unmarshal([]byte(rply.Data.(string)), &nymResp)
	if err != nil {
		return nil, err
	}

	//TODO: support multiple pubkeys
	txnTime := time.Unix(int64(rply.TxnTime), 0)
	verkey, _ := nymResp["verkey"].(string)
	pubKeyValue := base58.Decode(verkey)

	KID, err := localkms.CreateKID(pubKeyValue, kms.ED25519Type)
	if err != nil {
		return nil, err
	}

	pubKey := diddoc.NewPublicKeyFromBytes("#"+KID, keyType, "#id", pubKeyValue)
	verMethod := diddoc.NewReferencedVerificationMethod(pubKey, diddoc.Authentication, true)

	var svc []diddoc.Service
	serviceEndpoint, attrResp, err := r.getEndpoint(parsedDID.MethodSpecificID)
	if err == nil {
		s := diddoc.Service{
			ID:              "#agent",
			Type:            vdriapi.DIDCommServiceType,
			ServiceEndpoint: serviceEndpoint,
			Priority:        0,
			RecipientKeys:   []string{verkey},
		}

		svc = append(svc, s)
	}

	doc := &diddoc.Doc{
		Context:        []string{schemaV1},
		ID:             did,
		PublicKey:      []diddoc.PublicKey{*pubKey},
		Authentication: []diddoc.VerificationMethod{*verMethod},
		Service:        svc,
		Created:        &txnTime,
		Updated:        &txnTime,
	}

	end := time.Now()
	dur := end.Sub(start)
	out := &didResolution{
		Context:     context,
		DIDDocument: map[string]interface{}{},
		ResolverMetadata: map[string]interface{}{
			"driverId":  fmt.Sprintf(driverID, r.methodName),
			"driver":    driver,
			"retrieved": end,
			"duration":  dur.Milliseconds(),
		},
		MethodMetadata: map[string]interface{}{
			"nymResponse":  nymResp,
			"attrResponse": attrResp,
		},
	}

	d, _ := json.Marshal(doc)
	err = json.Unmarshal(d, &out.DIDDocument)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal doc")
	}
	return out, nil
}

func (r *HTTPIndyResolver) getEndpoint(did string) (string, map[string]interface{}, error) {
	rply, err := r.client.GetEndpoint(did)
	if err != nil || rply.Data == nil {
		return "", nil, errors.New("not found")
	}

	resp := map[string]interface{}{}
	err = json.Unmarshal([]byte(rply.Data.(string)), &resp)
	if err != nil {
		return "", nil, err
	}

	mm, ok := resp["endpoint"].(map[string]interface{})
	if !ok {
		return "", nil, errors.New("not found")
	}

	ep, ok := mm["endpoint"].(string)
	if !ok {
		return "", nil, errors.New("not found")
	}

	return ep, resp, nil
}
