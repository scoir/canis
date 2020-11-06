package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
	"github.com/makiuchi-d/gozxing"
	qr "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/mr-tron/base58"
	"github.com/skip2/go-qrcode"
)

const (
	URL               = "https://agency.ibmsecurity.verify-creds.com"
	Account           = "phil@scoir.com"
	AccountPassword   = "yc086z4wj8"
	AgentName         = "ibm-test-agent-for-canis"
	AgentPassword     = "canispw"
	DefaultAgentSeed  = "1111111111111111111111IBMInterop"
	DefaultLedgerName = "sovrin.staging"
)

type newAgentRequest struct {
	ID                string `json:"id"`
	Password          string `json:"password"`
	Seed              string `json:"seed"`
	DefaultLedgerName string `json:"default_ledger_name"`
}

func main() {

	deleteAgent()
	createAgent()
	generateInvitation()
}

func deleteAgent() {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/agents/%s", URL, AgentName), nil)
	if err != nil {
		log.Fatalln("unexpected error creating delete agent request", err)
	}
	req.SetBasicAuth(Account, AccountPassword)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalln("unexpected error calling delete agent", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("response code %d from delete agent\n", resp.StatusCode)
		return
	}

	log.Printf("agent %s successfully deleted\n", AgentName)
}

func createAgent() {
	agent := &newAgentRequest{
		ID:                AgentName,
		Password:          AgentPassword,
		Seed:              DefaultAgentSeed,
		DefaultLedgerName: DefaultLedgerName,
	}

	bits, err := json.Marshal(agent)
	if err != nil {
		log.Fatalln("unexpected error marshalling create agent request", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/agents", URL), bytes.NewBuffer(bits))
	if err != nil {
		log.Fatalln("unexpected error creating new agent request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(Account, AccountPassword)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalln("unexpected error calling new agent", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("response code %d from new agent\n", resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("unexpected error reading invitation body", err)
	}

	fmt.Println(string(b))

	var pubkey ed25519.PublicKey
	var privkey ed25519.PrivateKey
	privkey = ed25519.NewKeyFromSeed([]byte(DefaultAgentSeed))
	pubkey = privkey.Public().(ed25519.PublicKey)
	did, err := identifiers.CreateDID(&identifiers.MyDIDInfo{PublicKey: pubkey, Cid: true, MethodName: "sov"})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("New Agent DID:", did.String())
	fmt.Println("New Agent Verkey:", did.AbbreviateVerkey())
	fmt.Println("Place These in Wallet:")
	fmt.Println("Public:", base58.Encode(pubkey))
	fmt.Println("Private:", base58.Encode(privkey))
	log.Printf("agent %s successfully created\n", AgentName)
}

func generateInvitation() {
	buf := bytes.NewBuffer([]byte("{}"))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/invitations", URL), buf)
	if err != nil {
		log.Fatalln("unexpected error creating invitation request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(AgentName, AgentPassword)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Fatalln("unexpected error calling get invitation", err)
	}
	m := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		log.Fatalln("unexpected error decoding invitation request", err)
	}

	inviteURL := m["url"].(string)
	u, err := url.Parse(inviteURL)
	if err != nil {
		log.Fatalln("invalid invite URL", err)
	}

	invite := u.Query().Get("c_i")

	fname := "./ibm-verify-cred-invite.png"
	err = qrcode.WriteFile(invite, qrcode.Medium, 256, fname)
	if err != nil {
		log.Fatalln("unexpected error generating QR code", err)
	}

}

func encode() {
	resp, err := http.Get("http://local.scoir.com:7779/agents/hogwarts/invitation/subject")
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}
	defer resp.Body.Close()

	m := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		log.Fatalln("unexpected error decoding invitation request", err)
	}

	b := m["invitation"].(string)

	ci := base64.URLEncoding.EncodeToString([]byte(b))
	str := fmt.Sprintf("http://192.168.86.30/?c_i=%s", ci)

	fmt.Println(b)

	fname := "./invite.png"
	err = qrcode.WriteFile(str, qrcode.Medium, 256, fname)
	if err != nil {
		log.Fatal(err)
	}

}

func decode() {
	imgdata, err := ioutil.ReadFile("./verity-invite.png")
	if err != nil {
		log.Fatalln(err)
	}
	img, _, err := image.Decode(bytes.NewReader(imgdata))
	if err != nil {
		log.Fatalln(err)
	}
	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)

	// decode image
	qrReader := qr.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)

	fmt.Println(result)
}
