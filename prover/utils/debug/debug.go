//go:build debug

package debug

import "fmt"

// Try the condition and panic if it is does not satisfied
func Assert(cond func() bool, msg string, args ...any) {
	if !cond() {
		error_msg := fmt.Sprintf("Debug assertion failed: %v", msg)
		panic(fmt.Errorf(error_msg, args...))
	}
}

// Print a custom message, jump line by itself
func Printf(msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
}
