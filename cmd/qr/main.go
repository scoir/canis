package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/makiuchi-d/gozxing"
	qr "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/skip2/go-qrcode"
)

func main() {
	encode()
	//decode()
}

func encode() {
	req, err := http.NewRequest("GET", "http://local.scoir.com:7779/agents/highschool-agent-1/invitation/student-c", nil)
	if err != nil {
		log.Fatalln("unexpected error creating request", err)
	}
	req.Header.Set("X-API-Key", "D3YYdahdgC7VZeJwP4rhZcozCRHsqQT3VKxK9hTc2Yoh")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error requesting invitation from issuer: %v\n", err)
	}
	defer resp.Body.Close()

	m := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		log.Fatalln("struggled to decode response body, sigh", err)
	}

	b := m["invitation"].(string)

	//str := fmt.Sprintf("http://192.168.86.30/?c_i=%s", ci)

	fmt.Println(b)

	fname := "./invite.png"
	err = qrcode.WriteFile(b, qrcode.Medium, 256, fname)
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
