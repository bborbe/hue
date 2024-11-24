// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import "strings"

func ParseDeploymentsFromString(value string) Deployments {
	return ParseDeployments(strings.FieldsFunc(value, func(r rune) bool {
		return r == ','
	}))
}

func ParseDeployments(values []string) Deployments {
	result := make(Deployments, len(values))
	for i, value := range values {
		result[i] = Deployment(value)
	}
	return result
}

type Deployments []Deployment

func (c Deployments) Contains(deployment Deployment) bool {
	for _, o := range c {
		if o == deployment {
			return true
		}
	}
	return false
}

type Deployment string

func (d Deployment) String() string {
	return string(d)
}
