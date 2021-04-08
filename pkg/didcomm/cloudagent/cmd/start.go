/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/credential"
	"github.com/scoir/canis/pkg/didcomm/cloudagent"
	"github.com/scoir/canis/pkg/framework"
	canisproof "github.com/scoir/canis/pkg/presentproof"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm cloudagent service",
	Long:  `Starts a didcomm cloudagent service`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	i, err := cloudagent.New(ctx)
	if err != nil {
		log.Fatalln("unable to initialize mediator", err)
	}

	actx, _ := ctx.GetAriesContext()
	pofHandler, err := cloudagent.NewProofHandler(actx)
	if err != nil {
		log.Fatalln("unable to create proof handler", err)
	}

	prov := framework.NewSimpleProvider(actx)
	psup, err := canisproof.New(prov)
	if err != nil {
		log.Fatalln("unable to create proof supervisor", err)
	}

	err = psup.Start(pofHandler)
	if err != nil {
		log.Fatalln("unable to start proof supervisor", err)
	}

	credHandler, err := cloudagent.NewCredHandler(ctx)
	if err != nil {
		log.Fatalln("unable to create cred handler", err)
	}

	sup, err := credential.New(credHandler)
	if err != nil {
		log.Fatalln("unable to create credential supervisor for cloud agent", err)
	}

	err = sup.Start(credHandler)
	if err != nil {
		log.Fatalln(err, "unable to start credential supervisor for cloud agent")
	}

	err = i.Start()
	if err != nil {
		log.Println("cloudagent exited with", err)
	}

}

func init() {
	rootCmd.AddCommand(startCmd)
}
