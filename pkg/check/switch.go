// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

// NewSwitch returns main if fn returns true otherwise fallback
func NewSwitch(fn func() bool, main, fallback Check) Check {
	if fn() {
		return main
	}
	return fallback
}
