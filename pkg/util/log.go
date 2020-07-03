/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"log"
	"time"
)

func Logger(err error, _ time.Duration) {
	log.Println(err)
}
