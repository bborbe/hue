// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package k8s

import "strings"

func ParseStatefulSetsFromString(value string) StatefulSets {
	return ParseStatefulSets(strings.FieldsFunc(value, func(r rune) bool {
		return r == ','
	}))
}

func ParseStatefulSets(values []string) StatefulSets {
	result := make(StatefulSets, len(values))
	for i, value := range values {
		result[i] = StatefulSet(value)
	}
	return result
}

type StatefulSets []StatefulSet

func (c StatefulSets) Contains(statefulSet StatefulSet) bool {
	for _, o := range c {
		if o == statefulSet {
			return true
		}
	}
	return false
}

type StatefulSet string

func (s StatefulSet) String() string {
	return string(s)
}
