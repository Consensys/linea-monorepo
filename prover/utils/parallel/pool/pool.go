package pool

import (
	"runtime"
	"sync"
)

var queue chan func() = make(chan func())
var available chan struct{} = make(chan struct{}, runtime.GOMAXPROCS(0))
var once sync.Once
