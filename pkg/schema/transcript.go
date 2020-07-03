/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package schema

type TranscriptCredential struct {
	Name   string
	Date   string
	Degree string
	Age    string
}

func (r *TranscriptCredential) MarshalAttrs() []*Attr {
	var attrs []*Attr

	attrs = append(attrs, &Attr{Name: "name", Value: r.Name})
	attrs = append(attrs, &Attr{Name: "date", Value: r.Date})
	attrs = append(attrs, &Attr{Name: "degree", Value: r.Degree})
	attrs = append(attrs, &Attr{Name: "age", Value: r.Age})

	return attrs
}
