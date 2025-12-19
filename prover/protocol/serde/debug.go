package serde

import (
	"fmt"
	"reflect"
	"sync"
)

// DebugMode enables detailed offset logging.
// We default it to true to capture the crash logs immediately.
var DebugMode = true

var logMu sync.Mutex

func logTrace(op string, typeName string, fieldName string, offset int64, info string) {
	if !DebugMode {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	fmt.Printf("[%s] %-25s . %-20s | Off: %-6d | %s\n",
		op, typeName, fieldName, offset, info)
}

func logStructStart(op string, t reflect.Type, startOffset int64) {
	if !DebugMode {
		return
	}
	logMu.Lock()
	defer logMu.Unlock()
	fmt.Printf("--- %s STRUCT %-15s (Start: %d, GoSize: %d) ---\n",
		op, t.Name(), startOffset, t.Size())
}
