package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {

	h := sha256.New()

	h.Write([]byte{0xaa})
	s := h.Sum(nil)
	for i := 0; i < len(s); i++ {
		fmt.Printf("%x ", s[i])
	}
	fmt.Println("")

	h.Write([]byte{0xbb})
	s = h.Sum(nil)
	for i := 0; i < len(s); i++ {
		fmt.Printf("%x ", s[i])
	}
	fmt.Println("")

	h.Reset()
	h.Write([]byte{0xaa, 0xbb})
	s = h.Sum([]byte{0xff})
	for i := 0; i < len(s); i++ {
		fmt.Printf("%x ", s[i])
	}
	fmt.Println("")

}
