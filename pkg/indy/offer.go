package indy

type CredentialOffer struct {
	SchemaID            string `json:"schema_id"`
	CredDefID           string `json:"cred_def_id"`
	KeyCorrectnessProof string `json:"key_correctness_proof"`
	Nonce               string `json:"nonce"`
}
