package jsonld

import (
	"fmt"
	"log"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	docutil "github.com/hyperledger/aries-framework-go/pkg/doc/util"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/clr"
	"github.com/scoir/canis/pkg/credential"
)

type Engine struct {
	ctx     *ariescontext.Provider
	credsup *credential.Supervisor
	subject *didexchange.Connection
	offerID string
}

func (r *Engine) generateCredential() *verifiable.Credential {
	var issued = time.Date(2010, time.January, 1, 19, 23, 24, 0, time.UTC)

	record := &clr.CLR{
		Context: []string{
			"https://purl.imsglobal.org/spec/clr/v1p0/context/clr_v1p0.jsonld",
		},
		ID:   "did:scoir:abc123",
		Type: "Clr",
		Learner: &clr.Profile{
			ID:    "did:scoir:hss123",
			Type:  "Profile",
			Email: "student1@highschool.k12.edu",
		},
		Publisher: &clr.Profile{
			ID:    "did:scoir:highschool",
			Type:  "Profile",
			Email: "counselor@highschool.k12.edu",
		},
		Assertions: []*clr.Assertion{
			{
				ID:   "did:scoir:assert123",
				Type: "Assertion",
				Achievement: &clr.Achievement{
					ID:              "did:scoir:achieve123",
					AchievementType: "Achievement",
					Name:            "Mathmatics - Algebra Level 1",
				},
				IssuedOn: docutil.NewTime(issued),
			},
		},
		Achievements: nil,
		IssuedOn:     docutil.NewTime(issued),
	}

	vc := &verifiable.Credential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://purl.imsglobal.org/spec/clr/v1p0/context/clr_v1p0.jsonld",
		},
		ID: "http://example.edu/credentials/1872",
		Types: []string{
			"VerifiableCredential",
			"Clr"},
		Subject: record,
		Issuer: verifiable.Issuer{
			ID: r.subject.MyDID,
		},
		Issued:  docutil.NewTime(issued),
		Schemas: []verifiable.TypedID{},
		CustomFields: map[string]interface{}{
			"referenceNumber": 83294847,
		},
	}

	r.signCred(vc)
	return vc
}

func (r *Engine) signCred(vc *verifiable.Credential) {

	doc, err := r.ctx.VDRIRegistry().Resolve(r.subject.MyDID)
	if err != nil {
		log.Fatalln("unable to load my did doc")
	}

	signer, err := r.newCryptoSigner(doc.PublicKey[0].ID[1:])
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("ID", doc.PublicKey[0].ID)
	sigSuite := ed25519signature2018.New(
		suite.WithSigner(signer),
		suite.WithVerifier(ed25519signature2018.NewPublicKeyVerifier()))

	ldpContext := &verifiable.LinkedDataProofContext{
		SignatureType:           "Ed25519Signature2018",
		SignatureRepresentation: verifiable.SignatureProofValue,
		Suite:                   sigSuite,
		VerificationMethod:      fmt.Sprintf("%s%s", r.subject.MyDID, doc.PublicKey[0].ID),
	}

	err = vc.AddLinkedDataProof(ldpContext)
	if err != nil {
		log.Fatalln(err)
	}

}

func (r *Engine) newCryptoSigner(kid string) (*subtle.ED25519Signer, error) {
	priv, err := r.ctx.KMS().Get(kid)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find key set")
	}

	kh := priv.(*keyset.Handle)
	prim, err := kh.Primitives()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load signer primitives")
	}
	return prim.Primary.Primitive.(*subtle.ED25519Signer), nil

}

func TranscriptAccepted(id string) func(threadID string, ack *model.Ack) {

	return func(threadID string, ack *model.Ack) {
		fmt.Printf("Transcript Accepted: %s", id)
	}
}

func CredentialError(threadID string, err error) {
	log.Println("step 1... failed!", threadID, err)
}
