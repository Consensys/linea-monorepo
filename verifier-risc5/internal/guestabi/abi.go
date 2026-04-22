package guestabi

const (
	StatusBase = uintptr(0x80eff000)
	StatusSize = uintptr(1 << 12)
	StatusEnd  = StatusBase + StatusSize

	InputBase = uintptr(0x80f00000)
	InputSize = uintptr(1 << 20)
	InputEnd  = InputBase + InputSize

	QEMUTestBase = uintptr(0x00100000)

	Magic   = uint32(0x56524659) // "VRFY"
	Version = uint32(1)

	StatusMagic   = uint32(0x56535441) // "VSTA"
	StatusVersion = uint32(1)

	StatusCodeSuccess    = uint32(1)
	StatusCodeInputError = uint32(2)
	StatusCodeMismatch   = uint32(3)

	QEMUTestPass = uint32(0x5555)
	QEMUTestFail = uint32(0x3333)

	HeaderSize = uintptr(24)
	MaxWords   = int((InputSize - HeaderSize) / 8)
)

type Header struct {
	Magic     uint32
	Version   uint32
	WordCount uint32
	Reserved  uint32
	Expected  uint64
}

type Status struct {
	Magic    uint32
	Version  uint32
	Code     uint32
	Reserved uint32
	Result   uint64
	Expected uint64
}
