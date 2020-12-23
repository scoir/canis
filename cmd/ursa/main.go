package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/stretchr/testify/require"
)

func main() {
	t := &testing.T{}
	schemaBuilder, err := ursa.NewCredentialSchemaBuilder()
	require.NoError(t, err)
	err = schemaBuilder.AddAttr("attr1")
	require.NoError(t, err)
	schema, err := schemaBuilder.Finalize()
	require.NoError(t, err)

	nonSchemaBuilder, err := ursa.NewNonCredentialSchemaBuilder()
	require.NoError(t, err)
	err = nonSchemaBuilder.AddAttr("master_secret")
	require.NoError(t, err)
	nonSchema, err := nonSchemaBuilder.Finalize()

	credDef, err := ursa.NewCredentialDef(schema, nonSchema, false)
	require.NoError(t, err)
	require.NotNil(t, credDef)

	js, err := credDef.PubKey.ToJSON()
	require.NoError(t, err)
	fmt.Println(string(js))

	masterSecret, err := ursa.NewMasterSecret()
	require.NoError(t, err)
	js, err = masterSecret.ToJSON()
	require.NoError(t, err)
	m := struct {
		MS string `json:"ms"`
	}{}
	err = json.Unmarshal(js, &m)
	require.NoError(t, err)

	fmt.Println(m.MS)

	valuesBuilder, err := ursa.NewValueBuilder()
	require.NoError(t, err)
	err = valuesBuilder.AddDecHidden("master_secret", m.MS)
	require.NoError(t, err)
	err = valuesBuilder.AddDecKnown("attr1", "5944657099558967239210949258394887428692050081607692519917050011144233115103")
	require.NoError(t, err)

	values, err := valuesBuilder.Finalize()
	require.NoError(t, err)

	credentialNonce, err := ursa.NewNonce()
	require.NoError(t, err)

	js, err = credentialNonce.ToJSON()
	require.NoError(t, err)
	fmt.Println(string(js))

	blindedSecrets, err := ursa.BlindCredentialSecrets(credDef.PubKey, credDef.KeyCorrectnessProof, credentialNonce, values)
	require.NoError(t, err)

	js, err = blindedSecrets.Handle.ToJSON()
	require.NoError(t, err)
	fmt.Println(string(js))
	js, err = blindedSecrets.CorrectnessProof.ToJSON()
	require.NoError(t, err)
	fmt.Println(string(js))

}
