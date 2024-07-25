package wizard

import (
	"fmt"
	"strconv"
	"testing"
)

func TestTraceBack(t *testing.T) {

	a := func() string {
		return b()
	}

	fmt.Print("\n\n")
	a()
	fmt.Print("\n\n")

	t.FailNow()
}

func b() string {
	frames := getTraceBackFrames(0, 4)

	for _, frame := range frames {
		f, l := frame.File, frame.Line
		fmt.Print(f + ":" + strconv.Itoa(l) + "\n")
	}

	return "ok"
}
