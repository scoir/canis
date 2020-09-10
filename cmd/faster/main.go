package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	pdf "github.com/unidoc/unipdf/v3/model"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: go run pdf_extract_text.go input.pdf\n")
		os.Exit(1)
	}

	err := license.SetLicenseKey(`
-----BEGIN UNIDOC LICENSE KEY-----
eyJsaWNlbnNlX2lkIjoiNzA1YTBlMDktNzNmYS00NGQzLTU1NGYtZTg4NTMwYjkyOWNmIiwiY3VzdG9tZXJfaWQiOiJhMTVmMmQ3Ny00ZGFmLTRhZTEtNzJiMS1lYjY2NjEwYmUzYzUiLCJjdXN0b21lcl9uYW1lIjoiU2NvaXIiLCJjdXN0b21lcl9lbWFpbCI6Impvc2hAc2NvaXIuY29tIiwidGllciI6ImJ1c2luZXNzIiwiY3JlYXRlZF9hdCI6MTU5NjQ4Nzg1MiwiZXhwaXJlc19hdCI6MTYyODAzNTE5OSwiY3JlYXRvcl9uYW1lIjoiVW5pRG9jIFN1cHBvcnQiLCJjcmVhdG9yX2VtYWlsIjoic3VwcG9ydEB1bmlkb2MuaW8iLCJ1bmlwZGYiOnRydWUsInVuaW9mZmljZSI6ZmFsc2UsInRyaWFsIjpmYWxzZX0=
+
HErx4THd1LWpMlEdvgRy2Fp0pT5Bpm3mI43xClRvp4kCj0kzBfXneo0UaZNPyJGufawyiwcpE5uhMao0BX3hnUUwUI83YAP2eva52A2zK1fAVyCQXV7kT/A1NP5GG+LycXzCDvRdwoU8+LGHnJyg3Be8ZWtAMFtNnAKBY3sDCDvZDqltIlnw/mHJVu80qItKseZLPgoY1CyQFk8ziPNSevE7ci7UXh2v7p29HM8joyN4t+JS9ppMuY0DH/KdvkAkB53jKD2arQfwFNBl855imv57i8NauI8G8WH18AUZLO6cV61GBubktdQl8G73Vx5CvEdjT2E66UhzTsMA+aPx0w==
-----END UNIDOC LICENSE KEY-----
		`, "Scoir")

	if err != nil {
		log.Fatalln("bad license", err)
	}

	inputPath := os.Args[1]

	f, err := os.Open(inputPath)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		log.Fatalln(err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			log.Fatalln(err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Fatalln(err)
		}

		txt, err := ex.ExtractText()
		if err != nil {
			log.Fatalln(err)
		}

		if strings.Contains(txt, "COURSE INFORMATION") {
			pageText, _, _, err := ex.ExtractPageText()
			if err != nil {
				log.Fatalln(err)
			}

			tms := pageText.Marks()
			ar := tms.Elements()

			for _, tm := range ar {
				fmt.Printf("%d: %s\n", tm.Offset, tm.Text)
			}

		}

	}

}

func extractCourseInformation(lines []string) {
	var district int
	i := strings.Index(lines[0], "DISTRICT:")
	_, err := fmt.Sscanf(lines[0][i:], "DISTRICT: %d", &district)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("District:", district)
}

// outputPdfText prints out contents of PDF file to stdout.
func outputPdfText() error {
	inputPath := os.Args[1]

	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	fmt.Printf("--------------------\n")
	fmt.Printf("PDF to text extraction:\n")
	fmt.Printf("--------------------\n")
	pdfWriter := pdf.NewPdfWriter()
	count := 1
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return err
		}

		text, err := ex.ExtractText()
		if err != nil {
			return err
		}

		err = pdfWriter.AddPage(page)
		if err != nil {
			return err
		}

		if strings.Contains(text, "{END OF TRANSCRIPT}") {
			fWrite, err := os.Create(fmt.Sprintf("/tmp/pdf/%d.pdf", count))
			if err != nil {
				return err
			}

			err = pdfWriter.Write(fWrite)
			if err != nil {
				return err
			}
			err = fWrite.Close()
			if err != nil {
				return err
			}
			pdfWriter = pdf.NewPdfWriter()
			count++
		}

		//
		//fmt.Println("------------------------------")
		//fmt.Printf("Page %d:\n", pageNum)
		//fmt.Printf("\"%s\"\n", text)
		//fmt.Println("------------------------------")
	}

	return nil
}
