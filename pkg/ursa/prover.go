package ursa

/*
#cgo LDFLAGS: -L/usr/local/lib -lursa
#include "ursa_crypto.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/hyperledger/aries-framework-go/pkg/storage"
	"github.com/pkg/errors"

	"github.com/scoir/canis/pkg/indy/wrapper/vdr"
)

type Prover struct {
	store storage.Store
}

type provider interface {
	StorageProvider() storage.Provider
}

func NewProver(ctx provider) (*Prover, error) {
	s, err := ctx.StorageProvider().OpenStore("wallet")
	if err != nil {
		return nil, errors.Wrap(err, "unable to open store 'wallet'")
	}

	return &Prover{store: s}, nil
}

func (r *Prover) CreateMasterSecret(masterSecretID string) (string, error) {
	_, err := r.store.Get(masterSecretID)
	if err == nil {
		return "", errors.New("master secret id already exists")
	}

	var ms unsafe.Pointer
	result := C.ursa_cl_prover_new_master_secret(&ms)
	if result != 0 {
		return "", errors.Errorf("URSA error creating master secret: %d", result)
	}
	defer C.ursa_cl_master_secret_free(ms)

	var js *C.char
	result = C.ursa_cl_master_secret_to_json(ms, &js)
	if result != 0 {
		return "", errors.Errorf("URSA error converting master secret to json: %d", result)
	}
	defer C.free(unsafe.Pointer(js))

	str := C.GoString(js)

	val := struct {
		MS string `json:"ms"`
	}{}

	_ = json.Unmarshal([]byte(str), &val)
	fmt.Println(val.MS)
	err = r.store.Put(masterSecretID, []byte(val.MS))
	if err != nil {
		return "", errors.Wrap(err, "unable to store new master secret")
	}

	return masterSecretID, nil
}

type CredentialRequest struct {
	ProverDID                 string `json:"prover_did"`
	CredDefID                 string `json:"cred_def_id"`
	BlindedMS                 string `json:"blinded_ms"`
	BlindedMSCorrectnessProof string `json:"blinded_ms_correctness_proof"`
	Nonce                     string `json:"nonce"`
}

type CredentialRequestMetadata struct {
	MasterSecretBlindingData string `json:"master_secret_blinding_data"`
	Nonce                    string `json:"nonce"`
	MasterSecretName         string `json:"master_secret_name"`
}

func (r *Prover) CreateCredentialRequest(proverDID string, credDef *vdr.ClaimDefData, offer *CredentialOffer, masterSecretID string) (*CredentialRequest, *CredentialRequestMetadata, error) {
	var blindedCredentialSecrets, blindedCredentialSecretsCorrectnessProof, credentialSecretsBlindingFactors, nonce unsafe.Pointer

	val, err := r.store.Get(masterSecretID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "master secret not found")
	}

	var keyCorrectnessProof unsafe.Pointer
	credentialPubKey, err := CredDefHandle(credDef)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to build credential public key")
	}
	credValuesBuilder := NewValuesBuilder()
	credValuesBuilder.AddHiddenDec("master_secret", string(val)) //Could this not need encoding inside values builder?
	err = credValuesBuilder.Finalize()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to finalize builder")
	}

	kp := offer.KeyCorrectnessProof
	d, _ := json.Marshal(kp)
	keyCorrectnessProof, err = CredentialKeyCorrectnessProofFromJSON(string(d))
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get correctness proof from JSON")
	}

	nonce, err = NonceFromJSON(offer.Nonce)
	if err != nil {
		return nil, nil, err
	}

	result := C.ursa_cl_prover_blind_credential_secrets(
		credentialPubKey,
		keyCorrectnessProof,
		credValuesBuilder.Values(),
		nonce,
		&blindedCredentialSecrets,
		&credentialSecretsBlindingFactors,
		&blindedCredentialSecretsCorrectnessProof)
	if result != 0 {
		var errJson *C.char
		C.ursa_get_current_error(&errJson)
		defer C.free(unsafe.Pointer(errJson))
		return nil, nil, errors.Errorf("error from URSA blinding credential secrets: %s", C.GoString(errJson))
	}

	var blindedSecretsJson, proofJson, blindingFactorsJson *C.char
	result = C.ursa_cl_blinded_credential_secrets_correctness_proof_to_json(blindedCredentialSecretsCorrectnessProof, &proofJson)
	if result != 0 {
		var errJson *C.char
		C.ursa_get_current_error(&errJson)
		defer C.free(unsafe.Pointer(errJson))
		return nil, nil, errors.Errorf("error from URSA creating proof json: %s", C.GoString(errJson))
	}

	result = C.ursa_cl_blinded_credential_secrets_to_json(blindedCredentialSecrets, &blindedSecretsJson)
	if result != 0 {
		var errJson *C.char
		C.ursa_get_current_error(&errJson)
		defer C.free(unsafe.Pointer(errJson))
		return nil, nil, errors.Errorf("error from URSA creating blinded secrets json: %s", C.GoString(errJson))
	}

	result = C.ursa_cl_credential_secrets_blinding_factors_to_json(credentialSecretsBlindingFactors, &blindingFactorsJson)
	if result != 0 {
		var errJson *C.char
		C.ursa_get_current_error(&errJson)
		defer C.free(unsafe.Pointer(errJson))
		return nil, nil, errors.Errorf("error from URSA creating blinded secrets json: %s", C.GoString(errJson))
	}

	C.ursa_cl_blinded_credential_secrets_free(blindedCredentialSecrets)
	C.ursa_cl_blinded_credential_secrets_correctness_proof_free(blindedCredentialSecretsCorrectnessProof)
	C.ursa_cl_credential_secrets_blinding_factors_free(credentialSecretsBlindingFactors)

	cr := &CredentialRequest{
		ProverDID: proverDID,
	}

	cr.BlindedMS = C.GoString(blindedSecretsJson)
	cr.BlindedMSCorrectnessProof = C.GoString(proofJson)
	cr.Nonce, err = NewNonce()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create nonce for cred request")
	}

	md := &CredentialRequestMetadata{}
	md.MasterSecretBlindingData = C.GoString(blindingFactorsJson)

	return cr, md, nil
}
