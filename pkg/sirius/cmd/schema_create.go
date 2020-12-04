/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	api "github.com/scoir/canis/pkg/apiserver/api/protogen"
)

var version string
var schemaType string
var format string
var schemaCtx []string

var schemaCreateCmd = &cobra.Command{
	Use:   "create SCHEMA_NAME",
	Short: "Creates a schema.",
	RunE:  schemaCreate,
	Args:  cobra.ExactArgs(1),
}

func init() {
	schemaCmd.AddCommand(schemaCreateCmd)
	schemaCreateCmd.Flags().StringArrayVar(&schemaCtx, "context", []string{}, "list of context URLs this schema uses")
	schemaCreateCmd.Flags().StringArrayVar(&attrValues, "attr", []string{}, "list of attributes for this schema [NAME:TYPE] (can be repeated)")

	schemaCreateCmd.Flags().StringVar(&version, "version", "", "the schema version")
	_ = schemaCreateCmd.MarkFlagRequired("version")

	schemaCreateCmd.Flags().StringVar(&format, "format", "", "the schema format [hlindy-zkp-v1.0 | lds/ld-proof]")
	_ = schemaCreateCmd.MarkFlagRequired("format")
}

func schemaCreate(_ *cobra.Command, args []string) error {
	cli, err := ctx.GetAPIAdminClient()
	if err != nil {
		log.Fatalln("invalid server configuration", err)
	}

	ctx := context.Background()

	attrs := make([]*api.Attribute, len(attrValues))
	for i, value := range attrValues {
		vals := strings.Split(value, ":")
		if len(vals) != 2 {
			return errors.Errorf("invalid attribute %s, must be in format [name:TYPE]", value)
		}

		typ, ok := api.Attribute_Type_value[vals[1]]
		if !ok {
			return errors.Errorf("invalid attribute type, accepted values: [STRING | NUMBER | OBJECT | ARRAY | BOOL | NULL]", vals[1])
		}
		attrs[i] = &api.Attribute{
			Name: vals[0],
			Type: api.Attribute_Type(typ),
		}
	}

	schemaName := args[0]
	req := &api.CreateSchemaRequest{
		Schema: &api.NewSchema{
			Name:       schemaName,
			Version:    version,
			Type:       schemaType,
			Format:     format,
			Context:    schemaCtx,
			Attributes: attrs,
		},
	}

	_, err = cli.CreateSchema(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "unable to create schema %s", schemaName)
	}

	fmt.Printf("SCHEMA %s CREATED\n", schemaName)
	return nil
}
