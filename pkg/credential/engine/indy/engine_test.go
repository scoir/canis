package indy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/tink/go/signature/subtle"
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/decorator"
	kmsMock "github.com/hyperledger/aries-framework-go/pkg/mock/kms"
	"github.com/hyperledger/ursa-wrapper-go/pkg/libursa/ursa"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/hyperledger/indy-vdr/wrappers/golang/vdr"

	"github.com/scoir/canis/pkg/credential/engine/indy/mocks"
	"github.com/scoir/canis/pkg/datastore"
	gmock "github.com/scoir/canis/pkg/mock"
	"github.com/scoir/canis/pkg/schema"
)

func TestIssuerCredential(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}
		s := &datastore.Schema{
			ID: "schema-1",
			Attributes: []*datastore.Attribute{
				{Name: "attr1"},
			},
		}

		values := map[string]interface{}{
			"attr1": "test1234",
		}

		schemaBuilder, err := ursa.NewCredentialSchemaBuilder()
		require.NoError(t, err)

		for k := range values {
			err = schemaBuilder.AddAttr(k)
			require.NoError(t, err)
		}
		sch, err := schemaBuilder.Finalize()
		require.NoError(t, err)

		nonSchemaBuilder, err := ursa.NewNonCredentialSchemaBuilder()
		require.NoError(t, err)
		err = nonSchemaBuilder.AddAttr("master_secret")
		require.NoError(t, err)
		nonSchema, err := nonSchemaBuilder.Finalize()

		credDef, err := ursa.NewCredentialDef(sch, nonSchema, false)
		require.NoError(t, err)
		require.NotNil(t, credDef)

		pubKeyJS, err := credDef.PubKey.ToJSON()
		require.NoError(t, err)
		privKeyJS, err := credDef.PrivKey.ToJSON()
		require.NoError(t, err)

		ms, err := ursa.NewMasterSecret()
		require.NoError(t, err)
		js, err := ms.ToJSON()
		require.NoError(t, err)
		masterSecret := struct {
			MS string `json:"ms"`
		}{}
		err = json.Unmarshal(js, &masterSecret)
		require.NoError(t, err)

		valuesBuilder, err := ursa.NewValueBuilder()
		require.NoError(t, err)
		err = valuesBuilder.AddDecHidden("master_secret", masterSecret.MS)
		require.NoError(t, err)

		for k, v := range values {
			_, enc := ursa.EncodeValue(v)
			err = valuesBuilder.AddDecKnown(k, enc)
			require.NoError(t, err)
		}

		ursaValues, err := valuesBuilder.Finalize()
		require.NoError(t, err)

		credentialNonce, err := ursa.NewNonce()
		require.NoError(t, err)

		credentialNonceJS, err := credentialNonce.ToJSON()
		require.NoError(t, err)

		blindedSecrets, err := ursa.BlindCredentialSecrets(credDef.PubKey, credDef.KeyCorrectnessProof, credentialNonce, ursaValues)
		require.NoError(t, err)

		blindedMS, err := blindedSecrets.Handle.ToJSON()
		require.NoError(t, err)
		blindedSecretsCP, err := blindedSecrets.CorrectnessProof.ToJSON()
		require.NoError(t, err)

		//Need to track down the NONCES
		requestAttachment := decorator.AttachmentData{JSON: &datastore.CredentialRequest{
			ProverDID:                 "did:scr:123456789",
			CredDefID:                 "cred-def-1",
			BlindedMS:                 string(blindedMS),
			BlindedMSCorrectnessProof: string(blindedSecretsCP),
			Nonce:                     string(credentialNonceJS),
		}}

		offerID := "test-offer-id"
		offer := []byte(fmt.Sprintf(`{"cred_def_id": "test-creddef-id", "nonce": "%s"}`, strings.Trim(string(credentialNonceJS), "\"")))
		prov.store.On("Get", "test-offer-id").Return(offer, nil).Once()
		prov.store.On("Get", "test-creddef-id").Return([]byte(fmt.Sprintf(`{"privatekey": %s}`, privKeyJS)), nil).Once()

		pubKey := map[string]interface{}{}
		err = json.Unmarshal(pubKeyJS, &pubKey)
		require.NoError(t, err)

		prov.vdr.On("GetCredDef", "test-creddef-id").Return(&vdr.ReadReply{Data: map[string]interface{}{
			"primary": pubKey["p_key"],
		}}, nil)

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.NoError(t, err)
		require.NotNil(t, attachment)
	})
	t.Run("invalid attachment", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{}
		offerID := "test-offer-id"
		var values map[string]interface{}
		s := &datastore.Schema{}
		requestAttachment := decorator.AttachmentData{}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
	})

	t.Run("bad offer ID", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{}
		offerID := "test-offer-id"
		prov.store.On("Get", "test-offer-id").Return(nil, errors.New("not found"))

		var values map[string]interface{}
		s := &datastore.Schema{}
		requestAttachment := decorator.AttachmentData{JSON: &schema.IndyCredentialOffer{}}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
	})

	t.Run("bad cred def ID", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{}
		offerID := "test-offer-id"
		prov.store.On("Get", "test-offer-id").Return([]byte(`{"cred_def_id": "test-creddef-id"}`), nil).Once()
		prov.store.On("Get", "test-creddef-id").Return(nil, errors.New("not found")).Once()
		prov.vdr.On("GetCredDef", "test-creddef-id").Return(&vdr.ReadReply{Data: map[string]interface{}{
			"primary": map[string]interface{}{},
		}}, nil)

		var values map[string]interface{}
		s := &datastore.Schema{}
		requestAttachment := decorator.AttachmentData{JSON: &schema.IndyCredentialOffer{}}

		attachment, err := engine.IssueCredential(issuerDID, s, offerID, requestAttachment, values)
		require.Nil(t, attachment)
		require.Error(t, err)
	})

}

func TestRegisterSchema(t *testing.T) {
	t.Run("bas schema external id", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)
		require.NotNil(t, engine)

		registrantDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}
		s := &datastore.Schema{
			ID:               "schema-1",
			ExternalSchemaID: "schema-external-id",
			Attributes: []*datastore.Attribute{
				{Name: "attr1"},
			},
		}

		prov.vdr.On("GetSchema", "schema-external-id").Return(nil, errors.New("not found"))

		err = engine.RegisterSchema(registrantDID, s)
		require.Error(t, err)
	})
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)
		require.NotNil(t, engine)

		registrantDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: "test",
			},
		}

		s := &datastore.Schema{
			ID:               "schema-1",
			ExternalSchemaID: "schema-external-id",
			Attributes: []*datastore.Attribute{
				{Name: "attr1"},
			},
		}

		rply := &vdr.ReadReply{SeqNo: 23}
		prov.vdr.On("GetSchema", "schema-external-id").Return(rply, nil)

		kh, err := kmsMock.CreateMockED25519KeyHandle()
		require.NoError(t, err)
		prim, err := kh.Primitives()
		require.NoError(t, err)
		mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)
		prov.kms.GetKeyValue = kh

		prov.vdr.On("CreateClaimDef", "123456789", uint32(23), mock.AnythingOfType("map[string]interface {}"), map[string]interface{}(nil), mysig).Return("cred-def-id", nil)
		prov.store.On("Put", "cred-def-id", mock.AnythingOfType("[]uint8")).Return(nil)
		err = engine.RegisterSchema(registrantDID, s)
		require.NoError(t, err)
	})
}

func TestAccept(t *testing.T) {
	prov := newProvider()

	engine, err := New(prov.provider)
	require.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		ac := engine.Accept("hlindy-zkp-v1.0")
		require.True(t, ac)
	})
	t.Run("sad", func(t *testing.T) {
		ac := engine.Accept("lds/ld-proof")
		require.False(t, ac)
	})
}

func TestCreateSchema(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)
		issuer := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: "test",
			},
		}
		s := &datastore.Schema{
			Name:    "schema-name",
			Version: "1.2",
			Attributes: []*datastore.Attribute{
				{
					Name: "field1",
					Type: 0,
				},
				{
					Name: "field2",
					Type: 0,
				},
			},
		}

		kh, err := kmsMock.CreateMockED25519KeyHandle()
		require.NoError(t, err)
		prim, err := kh.Primitives()
		require.NoError(t, err)
		mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)

		prov.kms.GetKeyValue = kh

		prov.vdr.On("GetSchema", "123456789:2:schema-name:1.2").Return(nil, errors.New("not found"))
		prov.vdr.On("CreateSchema", issuer.DID.MethodID(), "schema-name", "1.2", []string{"field1", "field2"}, mysig).Return("test-schema-id", nil)

		sid, err := engine.CreateSchema(issuer, s)
		require.NoError(t, err)
		require.Equal(t, "test-schema-id", sid)

	})
	t.Run("existing schema", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)
		issuer := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: "test",
			},
		}
		s := &datastore.Schema{}

		prov.vdr.On("GetSchema", "123456789:2::").Return(&vdr.ReadReply{SeqNo: 1234}, nil)

		sid, err := engine.CreateSchema(issuer, s)
		require.NoError(t, err)
		require.Equal(t, "123456789:2::", sid)

	})
	t.Run("no keypair found", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)
		issuer := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: "test",
			},
		}
		s := &datastore.Schema{}

		prov.vdr.On("GetSchema", "123456789:2::").Return(nil, errors.New("not found"))

		prov.kms.GetKeyErr = errors.New("not found")

		sid, err := engine.CreateSchema(issuer, s)
		require.Error(t, err)
		require.Empty(t, sid)

	})
	t.Run("vdr failure", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)
		issuer := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
			KeyPair: &datastore.KeyPair{
				ID:        "123",
				PublicKey: "test",
			},
		}
		s := &datastore.Schema{
			Name:    "schema-name",
			Version: "1.2",
			Attributes: []*datastore.Attribute{
				{
					Name: "field1",
					Type: 0,
				},
				{
					Name: "field2",
					Type: 0,
				},
			},
		}

		kh, err := kmsMock.CreateMockED25519KeyHandle()
		require.NoError(t, err)
		prim, err := kh.Primitives()
		require.NoError(t, err)
		mysig := prim.Primary.Primitive.(*subtle.ED25519Signer)

		prov.kms.GetKeyValue = kh
		prov.vdr.On("GetSchema", "123456789:2:schema-name:1.2").Return(nil, errors.New("not found"))
		prov.vdr.On("CreateSchema", issuer.DID.MethodID(), "schema-name", "1.2", []string{"field1", "field2"}, mysig).Return("", errors.New("BOOM"))

		sid, err := engine.CreateSchema(issuer, s)
		require.Error(t, err)
		require.Empty(t, sid)

	})
}

func TestGetSchemaForProposal(t *testing.T) {
	t.Run("get schema", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		proposal := []byte(`{"schema_id": "123"}`)
		schemaID, err := engine.GetSchemaForProposal(proposal)
		require.NoError(t, err)
		require.Equal(t, "123", schemaID)
	})
	t.Run("get schema - bad JSON", func(t *testing.T) {
		prov := newProvider()

		engine, err := New(prov.provider)
		require.NoError(t, err)

		proposal := []byte(`{"schema_id": "`)
		schemaID, err := engine.GetSchemaForProposal(proposal)
		require.Error(t, err)
		require.Equal(t, "", schemaID)
	})
}

func TestCreateCredentialOffer(t *testing.T) {
	t.Run("bad schema id", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}

		s := &datastore.Schema{
			ID:               "schema-id",
			ExternalSchemaID: "abc",
			Attributes: []*datastore.Attribute{
				{
					Name: "attr1",
				},
			},
		}

		prov.vdr.On("GetSchema", "abc").Return(nil, errors.New("no schema"))

		offerID, offer, err := engine.CreateCredentialOffer(issuerDID, "", s, nil)
		require.Error(t, err)
		require.Empty(t, offerID)
		require.Nil(t, offer)
	})
	t.Run("wallet entry not found", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}

		s := &datastore.Schema{
			ID:               "schema-id",
			ExternalSchemaID: "abc",
			Attributes: []*datastore.Attribute{
				{
					Name: "attr1",
				},
			},
		}

		indySchema := &vdr.ReadReply{SeqNo: 1}
		prov.vdr.On("GetSchema", "abc").Return(indySchema, nil)
		prov.store.On("Get", "123456789:3:CL:1:default").Return(nil, errors.New("wallet not found"))

		offerID, offer, err := engine.CreateCredentialOffer(issuerDID, "", s, nil)
		require.Error(t, err)
		require.Empty(t, offerID)
		require.Nil(t, offer)

	})
	t.Run("nonce failure", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}

		s := &datastore.Schema{
			ID:               "schema-id",
			ExternalSchemaID: "abc",
			Attributes: []*datastore.Attribute{
				{
					Name: "attr1",
				},
			},
		}

		rec := &creddefWalletRecord{
			KeyCorrectnessProof: getKeyCorrectnessProof(t),
		}

		recData, err := json.Marshal(rec)
		require.NoError(t, err)

		indySchema := &vdr.ReadReply{SeqNo: 1}
		prov.vdr.On("GetSchema", "abc").Return(indySchema, nil)
		prov.store.On("Get", "123456789:3:CL:1:default").Return(recData, nil)
		prov.oracle.On("NewNonce").Return("", errors.New("no nonce for you"))

		offerID, offer, err := engine.CreateCredentialOffer(issuerDID, "", s, nil)
		require.Error(t, err)
		require.Empty(t, offerID)
		require.Nil(t, offer)

	})
	t.Run("put fails", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}

		s := &datastore.Schema{
			ID:               "schema-id",
			ExternalSchemaID: "abc",
			Attributes: []*datastore.Attribute{
				{
					Name: "attr1",
				},
			},
		}

		rec := &creddefWalletRecord{
			KeyCorrectnessProof: getKeyCorrectnessProof(t),
		}

		recData, err := json.Marshal(rec)
		require.NoError(t, err)

		indySchema := &vdr.ReadReply{SeqNo: 1}
		prov.vdr.On("GetSchema", "abc").Return(indySchema, nil)
		prov.store.On("Get", "123456789:3:CL:1:default").Return(recData, nil)
		prov.oracle.On("NewNonce").Return("0987654321234567890", nil)
		prov.store.On("Put", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(errors.New("can't save"))

		offerID, offer, err := engine.CreateCredentialOffer(issuerDID, "", s, nil)
		require.Error(t, err)
		require.Empty(t, offerID)
		require.Nil(t, offer)

	})
	t.Run("happy path", func(t *testing.T) {
		prov := newProvider()
		defer prov.Asserts(t)

		engine, err := New(prov.provider)
		require.NoError(t, err)

		issuerDID := &datastore.DID{
			DID: &identifiers.DID{
				DIDVal: identifiers.DIDValue{
					MethodSpecificID: "123456789",
					Method:           "scr",
				},
			},
		}

		s := &datastore.Schema{
			ID:               "schema-id",
			ExternalSchemaID: "abc",
			Attributes: []*datastore.Attribute{
				{
					Name: "attr1",
				},
			},
		}

		rec := &creddefWalletRecord{
			KeyCorrectnessProof: getKeyCorrectnessProof(t),
		}

		recData, err := json.Marshal(rec)
		require.NoError(t, err)

		indySchema := &vdr.ReadReply{SeqNo: 1}
		prov.vdr.On("GetSchema", "abc").Return(indySchema, nil)
		prov.store.On("Get", "123456789:3:CL:1:default").Return(recData, nil)
		prov.oracle.On("NewNonce").Return("0987654321234567890", nil)
		prov.store.On("Put", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

		offerID, offer, err := engine.CreateCredentialOffer(issuerDID, "", s, nil)
		require.NoError(t, err)
		require.NotEmpty(t, offerID)

		d, err := base64.StdEncoding.DecodeString(offer.Base64)
		require.NoError(t, err)

		m := map[string]interface{}{}
		err = json.Unmarshal(d, &m)

		v, ok := m["schema_id"].(string)
		require.True(t, ok)
		require.Equal(t, "abc", v)
		v, ok = m["cred_def_id"].(string)
		require.True(t, ok)
		require.Equal(t, "123456789:3:CL:1:default", v)
		v, ok = m["nonce"].(string)
		require.True(t, ok)
		require.Equal(t, "0987654321234567890", v)
	})
}

type provider struct {
	provider *MockProvider
	oracle   *mocks.Oracle
	vdr      *mocks.VDRClient
	sp       *gmock.MockProvider
	store    *gmock.MockStore
	kms      *kmsMock.KeyManager
}

func newProvider() provider {
	p := provider{
		provider: &MockProvider{},
		oracle:   &mocks.Oracle{},
		vdr:      &mocks.VDRClient{},
		store:    &gmock.MockStore{},
		sp:       &gmock.MockProvider{},
		kms:      &kmsMock.KeyManager{},
	}

	p.provider.On("IndyVDR").Return(p.vdr, nil)
	p.provider.On("Oracle").Return(p.oracle)
	p.provider.On("KMS").Return(p.kms)
	p.provider.On("StorageProvider").Return(p.sp)
	p.sp.On("OpenStore", "indy_engine").Return(p.store, nil)

	return p
}

func (r *provider) Asserts(t *testing.T) {
	r.provider.AssertExpectations(t)
	r.oracle.AssertExpectations(t)
	r.vdr.AssertExpectations(t)
	r.store.AssertExpectations(t)
	r.sp.AssertExpectations(t)
}

func getKeyCorrectnessProof(t *testing.T) map[string]interface{} {
	js := `{"c":"50072181238590910056344238666398725466118888516923203087905874091669173353742","xz_cap":"878805991126751213055044521696398886937697792914425096620064288696748506393245555450272809576471759224980622075377984711651926589356602633264628678416190803248865298953608918262970213411722926903034946745111405131839081760327586211150656414339688945880788836636058549803330522231610407609065886806335243159496186304720588497497004634227841539481977853721245891987221842979772305400456616096748866007418956951927609930827770747852149210703740422028955276489096539802470914836632736252142830945641028836552735077166636606388665104719176907265672640159316728969220101568652746675474181594956448417932423611850802505554446495228957031379636742216902577659188261766410209825614298842685819971343665","xr_cap":[["master_secret","1079142862517597231041106756438722786189565721805737488656067585792046295533414757536314556893561747506108795812631572246655333455604848566606056954357206632011558027487836398648364182432928540318064797514677441769074764635061186053355814943293673188899480156304117640300536667327008762963346897886748422237242230201605312919519681724227683641665693007132893761402201264865487125905211938879107409693514344233071431400172856885448873345194810060605634241068982787195077356078687475477177059427639375391779473194883264727633817707701401803338279853617397378178694218984511441009832985619089522001473924939716691056629880281175872338434506403219664040230483126571300572108829727592036341669550002"],["attr1","1152863606886242490750819397497401974001549007015330832959796021671787502231411582728336958672642961458877209300050491475063787835086659063432496517290214648176560555401143381513105789552473807047380673200560239761631112523315772518777606504282036191730229085675945212119683833905521721973244248030356603224803390501994298575218678514555230151808574807796338193378321114389685177747457781046015950541571549001513361084788011178394720864729200598384054489663205501284818561663716806701162846316526180072664921610247989450524366688150307813443839773366432070038974000015725957013357313236132513576851592979521650599177467110881116348269594508094770141721164078739043800360909336865543783540827825"]]}`
	out := map[string]interface{}{}
	err := json.Unmarshal([]byte(js), &out)
	require.NoError(t, err)
	return out
}
