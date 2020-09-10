package vdr

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type GetClaimDef struct {
	Operation     `json:",inline"`
	Origin        string `json:"origin"`
	SignatureType string `json:"signature_type"`
	Ref           uint32 `json:"ref"`
	Tag           string `json:"tag,omitempty"`
}

type ClaimDef struct {
	Operation     `json:",inline"`
	SignatureType string       `json:"signature_type"`
	Ref           uint32       `json:"ref"`
	Tag           string       `json:"tag,omitempty"`
	Data          ClaimDefData `json:"data"`
}

type ClaimDefData struct {
	ID         string                 `json:"-"`
	Primary    map[string]interface{} `json:"primary"`
	Revocation map[string]interface{} `json:"revocation,omitempty"`
}

func (r *ClaimDefData) PKey() string {
	d, _ := json.MarshalIndent(r.Primary, " ", " ")
	return string(d)
}

func (r *ClaimDefData) RKey() string {
	d, _ := json.MarshalIndent(r.Revocation, " ", " ")
	return string(d)
}

func NewGetClaimDef(origin string, ref uint32) *Request {
	return &Request{
		Operation: GetClaimDef{
			Operation:     Operation{Type: GET_CLAIM_DEF},
			Origin:        origin,
			SignatureType: "CL",
			Ref:           ref,
			Tag:           "default",
		},
		ProtocolVersion: protocolVersion,
		ReqID:           uuid.New().ID(),
	}
}

func NewClaimDef(from string, ref uint32, primary, revocation map[string]interface{}) *Request {
	return &Request{
		Operation: ClaimDef{
			Operation:     Operation{Type: CLAIM_DEF},
			SignatureType: "CL",
			Ref:           ref,
			Tag:           "default",
			Data:          ClaimDefData{Primary: primary, Revocation: revocation},
		},
		Identifier:      from,
		ProtocolVersion: protocolVersion,
		ReqID:           uuid.New().ID(),
	}
}

func (r *Client) CreateClaimDef(from string, ref uint32, pubKey, revocation map[string]interface{}, signer Signer) (string, error) {
	claimDef := NewClaimDef(from, ref, pubKey, revocation)

	resp, err := r.SubmitWrite(claimDef, signer)
	if err != nil {
		return "", errors.Wrap(err, "unable to create claim def")
	}

	return resp.TxnMetadata.TxnID, nil
}
