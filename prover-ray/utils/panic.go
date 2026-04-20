package utils

import "fmt"

func Panic(msg string, args ...any) {
	panic(fmt.Sprintf(msg, args...))
}
