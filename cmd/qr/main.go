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

	"github.com/makiuchi-d/gozxing"
	qr "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/skip2/go-qrcode"
)

func main() {

	d := []byte{85, 110, 115, 117, 112, 112, 111, 114, 116, 101, 100, 32, 67, 111, 110, 116, 101, 110, 116, 45, 116, 121, 112, 101, 32, 34, 97, 112, 112, 108, 105, 99, 97, 116, 105, 111, 110, 47, 115, 115, 105, 45, 97, 103, 101, 110, 116, 45, 119, 105, 114, 101, 34, 10}
	fmt.Println(string(d))

	//encode()
	//decode()
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
		log.Fatalln("fuck off", err)
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
