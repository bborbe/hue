// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import "strings"

func ParseContainersFromString(value string) Containers {
	return ParseContainers(strings.FieldsFunc(value, func(r rune) bool {
		return r == ','
	}))
}

func ParseContainers(values []string) Containers {
	result := make(Containers, len(values))
	for i, value := range values {
		result[i] = Container(value)
	}
	return result
}

type Containers []Container

func (c Containers) Contains(container Container) bool {
	for _, o := range c {
		if o == container {
			return true
		}
	}
	return false
}

type Container string

func (c Container) String() string {
	return string(c)
}
