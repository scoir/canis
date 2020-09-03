package issuer

import (
	"fmt"
	"log"
	"time"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/client/didexchange"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/model"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	icprotocol "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/issuecredential"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	docutil "github.com/hyperledger/aries-framework-go/pkg/doc/util"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	ariescontext "github.com/hyperledger/aries-framework-go/pkg/framework/context"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/clr"
	"github.com/scoir/canis/pkg/credential"
)

type credHandler struct {
	ctx     *ariescontext.Provider
	credsup *credential.Supervisor
	subject *didexchange.Connection
	offerID string
}

type prop interface {
	MyDID() string
	TheirDID() string
}

func (r *credHandler) ProposeCredentialMsg(_ service.DIDCommAction, _ *icprotocol.ProposeCredential) {
	panic("implement me")
}

func (r *credHandler) OfferCredentialMsg(_ service.DIDCommAction, _ *icprotocol.OfferCredential) {
	panic("implement me")
}

func (r *credHandler) IssueCredentialMsg(_ service.DIDCommAction, _ *icprotocol.IssueCredential) {
	panic("implement me")
}

func (r *credHandler) RequestCredentialMsg(e service.DIDCommAction, request *icprotocol.RequestCredential) {
	props := e.Properties.(prop)
	theirDID := props.TheirDID()

	if theirDID != r.subject.TheirDID {
		log.Println("invalid request for credential by", theirDID)
		return
	}

	for _, attach := range request.RequestsAttach {
		cred, _ := attach.Data.JSON.(map[string]interface{})
		id, _ := cred["id"].(string)
		if id == "" {
			log.Println("no ID found in request attachment")
			continue
		}

		var msg *icprotocol.IssueCredential
		thid, _ := e.Message.ThreadID()

		fmt.Printf("offerID: %s, credID: %s, threadID: %s\n", r.offerID, id, thid)

		if r.offerID == thid {
			msg = &icprotocol.IssueCredential{
				Type:    icprotocol.IssueCredentialMsgType,
				Comment: fmt.Sprintf("CLR Transcript"),
				CredentialsAttach: []decorator.Attachment{
					{Data: decorator.AttachmentData{JSON: r.generateCredential()}},
				},
			}

			//TODO:  Shouldn't this be built into the Supervisor??
			log.Println("setting up monitoring for", thid)
			mon := credential.NewMonitor(r.credsup)
			mon.WatchThread(thid, TranscriptAccepted(id), CredentialError)
		}
		e.Continue(icprotocol.WithIssueCredential(msg))

	}
}

func (r *credHandler) generateCredential() *verifiable.Credential {
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

func (r *credHandler) signCred(vc *verifiable.Credential) {

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

func (r *credHandler) newCryptoSigner(kid string) (*subtle.ED25519Signer, error) {
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
