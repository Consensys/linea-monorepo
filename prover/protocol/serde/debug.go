package serde

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
)

// TOGGLE THIS TO FALSE WHEN DONE
const DebugEnabled = true

var debugDepth int32

// traceEnter logs the start of a function/operation with indentation
func traceEnter(op string, v reflect.Value, extra ...any) {
	if !DebugEnabled {
		return
	}
	d := atomic.AddInt32(&debugDepth, 1)
	pad := strings.Repeat("|  ", int(d-1))

	typeStr := "nil"
	if v.IsValid() {
		typeStr = v.Type().String()
	}

	ptrInfo := ""
	if v.IsValid() && v.Kind() == reflect.Ptr {
		ptrInfo = fmt.Sprintf(" [Ptr: %p]", v.Interface())
	}

	fmt.Printf("%s-> %s: Type=%s%s %v\n", pad, op, typeStr, ptrInfo, extra)
}

// traceExit logs the end of a function
func traceExit(op string, err error, extra ...any) {
	if !DebugEnabled {
		return
	}
	d := atomic.LoadInt32(&debugDepth)
	pad := strings.Repeat("|  ", int(d-1))

	status := "OK"
	if err != nil {
		status = fmt.Sprintf("ERR: %v", err)
	}

	fmt.Printf("%s<- %s [%s] %v\n", pad, op, status, extra)
	atomic.AddInt32(&debugDepth, -1)
}

// traceLog logs an event at the current depth (e.g., writing a specific field)
func traceLog(msg string, args ...any) {
	if !DebugEnabled {
		return
	}
	d := atomic.LoadInt32(&debugDepth)
	pad := strings.Repeat("|  ", int(d))
	fmt.Printf("%s[LOG] "+msg+"\n", append([]any{pad}, args...)...)
}

// traceOffset logs exactly where we are in the file
func traceOffset(component string, offset int64) {
	if !DebugEnabled {
		return
	}
	d := atomic.LoadInt32(&debugDepth)
	pad := strings.Repeat("|  ", int(d))
	fmt.Printf("%s[@OFFSET %d] %s\n", pad, offset, component)
}
