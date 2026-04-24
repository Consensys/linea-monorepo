//go:build tamago && riscv64 && tamago_libriscv

package goos

import "unsafe"

const (
	ramStartAddr    uint    = 0x80000000
	ramSizeBytes    uint    = 0x01200000
	stackOffset     uint    = 0x1000
	qemuTestBase    uintptr = 0x00100000
	qemuTestPass    uint32  = 0x00005555
	qemuTestFail    uint32  = 0x00003333
	idleForever     int64   = 1<<63 - 1
	initialRNGState uint64  = 0x243f6a8885a308d3
)

var (
	RamStart       uint = ramStartAddr
	RamSize        uint = ramSizeBytes
	RamStackOffset uint = stackOffset

	Bloc uintptr

	Exit   = guestExit
	Idle   = guestIdle
	ProcID = guestProcID
	Task   func(sp, mp, gp, fn unsafe.Pointer)

	nanotimeState int64
	rngState      uint64 = initialRNGState
)

func CPUInit()
func Hwinit0()
func Printk(c byte)
func writeFinisher(value uint32)

func InitRNG() {}

func GetRandomData(b []byte) {
	for i := range b {
		rngState ^= rngState << 7
		rngState ^= rngState >> 9
		rngState ^= rngState << 8
		b[i] = byte(rngState >> 56)
	}
}

func Nanotime() int64 {
	nanotimeState += 1000
	return nanotimeState
}

func Hwinit1() {
	Exit = guestExit
	Idle = guestIdle
	ProcID = guestProcID
}

func guestProcID() uint64 {
	return 0
}

//go:nosplit
func guestIdle(until int64) {
	if until == idleForever {
		guestExit(0)
	}
}

//go:nosplit
func guestExit(code int32) {
	value := qemuTestPass
	if code != 0 {
		value = (uint32(code) << 16) | qemuTestFail
	}

	writeFinisher(value)
	for {
	}
}
