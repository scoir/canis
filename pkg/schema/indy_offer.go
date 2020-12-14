package schema

type IndyCredentialOffer struct {
	SchemaID            string                 `json:"schema_id"`
	CredDefID           string                 `json:"cred_def_id"`
	KeyCorrectnessProof map[string]interface{} `json:"key_correctness_proof"`
	Nonce               string                 `json:"nonce"`
}
