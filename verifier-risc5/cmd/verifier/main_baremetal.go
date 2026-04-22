//go:build baremetal

package main

func main() {
	Result = Compute()
	announceBaremetal(Result)

	for {
	}
}
