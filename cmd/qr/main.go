package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/makiuchi-d/gozxing"
	qr "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/skip2/go-qrcode"
)

func main() {

	//b := []byte{85, 110, 115, 117, 112, 112, 111, 114, 116, 101, 100, 32, 67, 111, 110, 116, 101, 110, 116, 45, 116, 121, 112, 101, 32, 34, 97, 112, 112, 108, 105, 99, 97, 116, 105, 111, 110, 47, 115, 115, 105, 45, 97, 103, 101, 110, 116, 45, 119, 105, 114, 101, 34, 10}
	//fmt.Println(string(b))

	encode()
	//decode()
}

func encode() {
	req, err := http.NewRequest("GET", "http://34.72.71.135:7779/agents/ibm-test-agent/invitation/ibm-mobile-agent?name=IBM%20Mobile%20Agent", nil)
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}
	req.Header.Set("X-API-Key", "D3YYdahdgC7VZeJwP4rhZcozCRHsqQT3VKxK9hTc2Yoh")
	//body := `{"type": "oob"}`
	//req, err := http.NewRequest("POST", "https://agency.keith.ti.verify-creds.com/api/v1/invitations", strings.NewReader(body))
	//if err != nil {
	//	log.Fatalln("unexpected error creating request", err)
	//}
	//req.Header.Set("Content-Type", "application/json")
	//req.SetBasicAuth("ibm-test-agent-for-canis", "canispw")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalln("bad code", resp.StatusCode, string(body))
	}

	m := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		log.Fatalln("unable to decode body", err)
	}

	b := m["Invitation"].(string)

	m = map[string]interface{}{}
	err = json.Unmarshal([]byte(b), &m)
	if err != nil {
		log.Fatalln("blah", err)
	}

	d, _ := json.MarshalIndent(m, " ", " ")
	fmt.Println(string(d))

	ci := base64.StdEncoding.EncodeToString([]byte(b))
	str := fmt.Sprintf("http://34.72.71.135:7779/invitation?oob=%s", ci)

	fmt.Println(str)

	oobInvite := struct {
		URL string `json:"url"`
	}{
		URL: str,
	}

	oobJSON, err := json.Marshal(oobInvite)
	if err != nil {
		log.Fatalln("unexpected error marshalling invite", err)
	}

	out, err := os.Create("invite-for-ibm.json")
	if err != nil {
		log.Fatalln("can't create ibm invite json", err)
	}

	_, _ = out.WriteString(string(oobJSON))
	out.Close()

	fname := "./invite.png"
	err = qrcode.WriteFile(string(oobJSON), qrcode.Medium, 256, fname)
	if err != nil {
		log.Fatal(err)
	}

}

func decode() {
	imgdata, err := ioutil.ReadFile("./ibm-verify-cred-invite.png")
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
