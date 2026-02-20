package collection

import "fmt"

// Panic panics with a formatted message.
// This local copy avoids circular import with utils package.
func panic_(msg string, args ...any) {
	panic(fmt.Sprintf(msg, args...))
}
