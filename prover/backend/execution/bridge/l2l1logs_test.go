package bridge

import (
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {

	a := L2L1Topic0()
	fmt.Printf("%x\n", a)

	b := GetRollingHashUpdateTopic0()
	fmt.Printf("%x\n", b)
	panic("xxx")
}
