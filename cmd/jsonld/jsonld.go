/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/hyperledger/indy-vdr/wrappers/golang/identifiers"
)

func main() {
	canon()
}

func canon() {
	record := verifiable.Credential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://schema.org/docs/jsonldcontext.json",
		},
		ID: "did:scoir:abc123",
		Subject: map[string]interface{}{
			"@id":                          "did:peer:1zQmYpHUXTfuJSpmQjHHhp1KSA4C2KTwMuqzzSf63BHbQ3za",
			"educationalCredentialAwarded": "Potions Master",
		},
	}

	d, _ := json.MarshalIndent(record, " ", " ")
	out := map[string]interface{}{}
	err := json.Unmarshal(d, &out)
	if err != nil {
		log.Fatalln("broken", err)
	}

	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	doc3, err := proc.Normalize(out, options)
	if err != nil {
		log.Fatalln(err)
	}

	d, _ = json.MarshalIndent(doc3, " ", " ")
	fmt.Println(string(d))

}

func fromRDF() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	triples := `
		<http://example.com/Subj1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://example.com/Type> .
		<http://example.com/Subj1> <http://example.com/prop1> <http://example.com/Obj1> .
		<http://example.com/Subj1> <http://example.com/prop2> "Plain" .
		<http://example.com/Subj1> <http://example.com/prop2> "2012-05-12"^^<http://www.w3.org/2001/XMLSchema#date> .
		<http://example.com/Subj1> <http://example.com/prop2> "English"@en .
	`

	doc, err := proc.FromRDF(triples, options)
	if err != nil {
		log.Fatalln(err)
	}

	d, _ := json.MarshalIndent(doc, " ", " ")
	fmt.Println(string(d))

	doc2, err := proc.Flatten(doc, nil, options)
	if err != nil {
		log.Fatalln(err)
	}

	d, _ = json.MarshalIndent(doc2, " ", " ")
	fmt.Println(string(d))

	doc3, err := proc.Normalize(doc, options)
	if err != nil {
		log.Fatalln(err)
	}

	d, _ = json.MarshalIndent(doc3, " ", " ")
	fmt.Println(string(d))

	doc4, err := proc.Expand(doc, options)
	if err != nil {
		log.Fatalln(err)
	}

	d, _ = json.MarshalIndent(doc4, " ", " ")
	fmt.Println(string(d))

	doc5, err := proc.ToRDF(doc, options)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(doc5)

	//d, _ = json.MarshalIndent(doc5, " ", " ")
	//fmt.Println(string(d))

}

func flatten() {
	//err := indy.ResolveDID("PkygzecB8VwTf9jAMYKDrS")
	//log.Fatalln(err)
	didinfo := &identifiers.MyDIDInfo{
		DID:        "",
		Cid:        true,
		MethodName: "ioe",
	}

	did, err := identifiers.CreateDID(didinfo)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("DID:", did.String())
	fmt.Println("Verkey:", did.AbbreviateVerkey())

	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	doc := map[string]interface{}{
		"@context": []interface{}{
			map[string]interface{}{
				"name": "http://xmlns.com/foaf/0.1/name",
				"homepage": map[string]interface{}{
					"@id":   "http://xmlns.com/foaf/0.1/homepage",
					"@type": "@id",
				},
			},
			map[string]interface{}{
				"ical": "http://www.w3.org/2002/12/cal/ical#",
			},
		},
		"@id":           "http://example.com/speakers#Alice",
		"name":          "Alice",
		"homepage":      "http://xkcd.com/177/",
		"ical:summary":  "Alice Talk",
		"ical:location": "Lyon Convention Centre, Lyon, France",
	}

	flattenedDoc, err := proc.Normalize(doc, options)
	if err != nil {
		log.Println("Error when flattening JSON-LD document:", err)
		return
	}

	ld.PrintDocument("JSON-LD flattened doc", flattenedDoc)

}
