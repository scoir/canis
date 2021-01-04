/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/scoir/canis/pkg/controller"
	"github.com/scoir/canis/pkg/didcomm/verifier"
	"github.com/scoir/canis/pkg/framework"
	"github.com/scoir/canis/pkg/presentproof"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the didcomm service",
	Long:  `Starts a didcomm service`,
	Run:   runStart,
}

func runStart(_ *cobra.Command, _ []string) {
	prov := framework.NewSimpleProvider(ctx.actx)
	ppsup, err := presentproof.New(prov)
	if err != nil {
		log.Fatalln("unable to create new proof supervisor", err)
	}

	reg, err := ctx.GetPresentationEngineRegistry()
	if err != nil {
		log.Fatalln("unable to initialize proof engine registry", err)
	}

	handler := verifier.NewProofHandler(ctx.store, reg)
	err = ppsup.Start(handler)
	if err != nil {
		log.Fatalln("unable to start proof supervisor", err)
	}

	i, err := verifier.New(ctx)
	if err != nil {
		log.Fatalln("unable to initialize verifier", err)
	}

	runner, err := controller.New(ctx, i)
	if err != nil {
		log.Fatalln("unable to start didcomm-verifier", err)
	}

	err = runner.Launch()
	if err != nil {
		log.Fatalln("launch errored with", err)
	}

}

func init() {
	rootCmd.AddCommand(startCmd)
}
