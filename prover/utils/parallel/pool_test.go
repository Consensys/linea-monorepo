package parallel

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestPoolSize(t *testing.T) {
	run()
	fmt.Println(len(available))
	assert.Equal(t, runtime.GOMAXPROCS(0), len(available))
}
