//go:build baremetal && !qemu_virt && !tamago_sifive_u

package main

func announceBaremetal(uint64) {}
