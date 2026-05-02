//go:build baremetal

package main

// Bare-metal guests have nowhere to return after reporting status.
func haltForever() {
	for {
	}
}
