/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/datastore"
	"github.com/scoir/canis/pkg/indy/wrapper/identifiers"
)

var seed string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the public key in a canis credential hub",
	Long:  `Initialize the public key in a canis credential hub`,
	Run:   initCluster,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVar(&seed, "seed", "", "seed for public DID")
}

func initCluster(_ *cobra.Command, _ []string) {
	did, err := getPublicDID()
	if err == nil {
		fmt.Println("public DID already set to", did)
		return
	}
	if seed != "" {
		saveExistingPublicDID()
	} else {
		log.Fatalln("seed flag is required")
	}
}

func getPublicDID() (string, error) {
	ds, err := ctx.StorageProvider()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve datastore: (%w)", err)
	}

	didds, _ := ds.OpenStore("DID")
	did, err := didds.GetPublicDID()
	if err != nil {
		return "", fmt.Errorf("no public DID set: (%w)", err)
	}
	return did.DID.String(), nil
}

func saveExistingPublicDID() {
	did, keyPair, err := identifiers.CreateDID(&identifiers.MyDIDInfo{
		Seed:       seed,
		Cid:        true,
		MethodName: "scr",
	})
	if err != nil {
		log.Fatalln(err)
	}

	vdr, err := ctx.VDR()
	if err != nil {
		log.Fatalln("unable to get VDR", err)
	}

	_, err = vdr.GetNym(did.String())
	if err != nil {
		log.Fatalln("DID must be registered to be public", err)
	}

	ds, err := ctx.StorageProvider()
	if err != nil {
		log.Fatalln(err)
	}
	didds, _ := ds.OpenStore("DID")

	var d = &datastore.DID{
		DID: did,
		KeyPair: &datastore.KeyPair{
			PublicKey:  keyPair.PublicKey(),
			PrivateKey: keyPair.PrivateKey(),
		},
		Endpoint: "",
	}
	err = didds.InsertDID(d)
	if err != nil {
		log.Fatalln(err)
	}

	err = didds.SetPublicDID(did.String())
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("stored public DID as", did.String())
}
