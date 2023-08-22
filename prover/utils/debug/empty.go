//go:build !debug

package debug

func Assert(cond func() bool, msg string, args ...any) {}

// Print a custom message, jump line by itself
func Printf(msg string, args ...any) {}
