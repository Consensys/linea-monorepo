package utils

import "fmt"

func Panic(msg string, args ...any) {
	panic(fmt.Sprintf(msg, args...))
}

func Require(cond bool, msg string, args ...any) {
	if !cond {
		Panic(msg, args...)
	}
}
