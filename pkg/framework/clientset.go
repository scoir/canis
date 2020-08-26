/*
Copyright Scoir Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package framework

import "k8s.io/client-go/kubernetes"

type Clientset struct {
	*kubernetes.Clientset
	Namespace string
}
