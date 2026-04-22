//go:build !baremetal

package main

func main() {
	Result = Compute()
	println("verifier result", Result)
}
