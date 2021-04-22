package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/tink/go/signature/subtle"
)

type edge struct {
	CloudAgentID   string
	PublicKey      []byte
	PrivateKey     []byte
	NextKey        []byte
	NextPrivateKey []byte
}

func main() {

	e := registerCloudAgent()

	conns := listConnections(e)
	d, _ := json.MarshalIndent(conns, " ", " ")
	fmt.Println(string(d))

	connectToAgent(e)

	results := listCredentials(e)
	d, _ = json.MarshalIndent(results, " ", " ")
	fmt.Println(string(d))

	creds, ok := results["credentials"].([]interface{})

	if ok {
		for _, i := range creds {
			cred := i.(map[string]interface{})
			if cred["status"] == "offered" {
				fmt.Println("Credential", cred["comment"], "has been offered")
				acceptOffer(e, cred["credential_id"].(string))
			}
		}
	}

	prs := listProofRequests(e)
	d, _ = json.MarshalIndent(prs, " ", " ")
	fmt.Println(string(d))
}

func acceptOffer(e *edge, credentialID string) {
	d := []byte(`{}`)
	req, err := http.NewRequest("POST", "https://canis.scoir.ninja/cloudagents/credentials/"+credentialID, bytes.NewBuffer(d))
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}

	privKey := ed25519.PrivateKey(e.PrivateKey)
	signer, err := subtle.NewED25519SignerFromPrivateKey(&privKey)
	if err != nil {
		panic(err)
	}

	sig, err := signer.Sign(d)
	if err != nil {
		panic(err)
	}

	req.Header.Add("x-canis-cloud-agent-id", e.CloudAgentID)
	req.Header.Add("x-canis-cloud-agent-signature", base64.URLEncoding.EncodeToString(sig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("unable to accept credential %s with error %s\n", credentialID, string(b))
		return
	}

	fmt.Printf("Credential %s successfully accepted\n", credentialID)

}

type registration struct {
	PublicKey []byte `json:"public_key"`
	NextKey   []byte `json:"next_key"`
	Secret    string `json:"secret"`
}

func registerCloudAgent() *edge {

	e := &edge{}

	d, err := ioutil.ReadFile("./edge.json")
	if err == nil {
		err := json.Unmarshal(d, e)
		if err == nil {
			return e
		}
	}

	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)

	e.PublicKey = edPub
	e.PrivateKey = edPriv

	nedPub, nedPriv, err := ed25519.GenerateKey(rand.Reader)

	e.NextKey = nedPub
	e.NextPrivateKey = nedPriv

	reg := &registration{
		Secret:    "ArwXoACJgOleVZ2PY7kXn7rA0II0mHYDhc6WrBH8fDAc",
		PublicKey: e.PublicKey,
		NextKey:   e.NextKey,
	}

	w := &bytes.Buffer{}
	enc := json.NewEncoder(w)
	_ = enc.Encode(reg)

	req, err := http.NewRequest("POST", "https://canis.scoir.ninja/cloudagents", w)
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	m := map[string]interface{}{}
	_ = json.Unmarshal(b, &m)

	fmt.Println(string(b))

	e.CloudAgentID = m["cloud_agent_id"].(string)

	d, _ = json.MarshalIndent(e, " ", " ")
	_ = ioutil.WriteFile("./edge.json", d, os.ModePerm)

	return e
}

func connectToAgent(e *edge) {
	req, err := http.NewRequest("GET", "http://34.72.71.135:7779/agents/ibm-test-agent/invitation/test-student-123", nil)
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}
	req.Header.Set("X-API-Key", "D3YYdahdgC7VZeJwP4rhZcozCRHsqQT3VKxK9hTc2Yoh")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}
	defer resp.Body.Close()
	d, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(d))

	req, err = http.NewRequest("POST", "https://canis.scoir.ninja/cloudagents/invitation", bytes.NewBuffer(d))
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}

	privKey := ed25519.PrivateKey(e.PrivateKey)
	siger, err := subtle.NewED25519SignerFromPrivateKey(&privKey)
	if err != nil {
		panic(err)
	}

	sig, err := siger.Sign(d)
	if err != nil {
		panic(err)
	}

	req.Header.Add("x-canis-cloud-agent-id", e.CloudAgentID)
	req.Header.Add("x-canis-cloud-agent-signature", base64.URLEncoding.EncodeToString(sig))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	b, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Println(string(b))

}

func listConnections(e *edge) map[string]interface{} {
	d := []byte(`{}`)
	req, err := http.NewRequest("POST", "https://canis.scoir.ninja/cloudagents/connections", bytes.NewBuffer(d))
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}

	privKey := ed25519.PrivateKey(e.PrivateKey)
	signer, err := subtle.NewED25519SignerFromPrivateKey(&privKey)
	if err != nil {
		panic(err)
	}

	sig, err := signer.Sign(d)
	if err != nil {
		panic(err)
	}

	req.Header.Add("x-canis-cloud-agent-id", e.CloudAgentID)
	req.Header.Add("x-canis-cloud-agent-signature", base64.URLEncoding.EncodeToString(sig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	out := map[string]interface{}{}
	defer resp.Body.Close()
	d, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(d, &out)
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(d))
	}

	return out
}

func listCredentials(e *edge) map[string]interface{} {
	d := []byte(`{}`)
	req, err := http.NewRequest("POST", "https://canis.scoir.ninja/cloudagents/credentials", bytes.NewBuffer(d))
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}

	privKey := ed25519.PrivateKey(e.PrivateKey)
	signer, err := subtle.NewED25519SignerFromPrivateKey(&privKey)
	if err != nil {
		panic(err)
	}

	sig, err := signer.Sign(d)
	if err != nil {
		panic(err)
	}

	req.Header.Add("x-canis-cloud-agent-id", e.CloudAgentID)
	req.Header.Add("x-canis-cloud-agent-signature", base64.URLEncoding.EncodeToString(sig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	out := map[string]interface{}{}
	defer resp.Body.Close()
	d, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(d, &out)
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(d))
	}

	return out
}

func listProofRequests(e *edge) map[string]interface{} {
	d := []byte(`{}`)
	req, err := http.NewRequest("POST", "https://canis.scoir.ninja/cloudagents/proof_requests", bytes.NewBuffer(d))
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}

	privKey := ed25519.PrivateKey(e.PrivateKey)
	signer, err := subtle.NewED25519SignerFromPrivateKey(&privKey)
	if err != nil {
		panic(err)
	}

	sig, err := signer.Sign(d)
	if err != nil {
		panic(err)
	}

	req.Header.Add("x-canis-cloud-agent-id", e.CloudAgentID)
	req.Header.Add("x-canis-cloud-agent-signature", base64.URLEncoding.EncodeToString(sig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}

	out := map[string]interface{}{}
	defer resp.Body.Close()
	d, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(d, &out)
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(d))
	}

	return out
}
