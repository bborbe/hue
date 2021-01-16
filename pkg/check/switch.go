package check

// NewSwitch returns main if fn returns true otherwise fallback
func NewSwitch(fn func() bool, main, fallback Check) Check {
	if fn() {
		return main
	}
	return fallback
}
