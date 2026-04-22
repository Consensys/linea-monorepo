//go:build baremetal && tamago_sifive_u

package main

func announceBaremetal(value uint64) {
	println("verifier result", value)
}
