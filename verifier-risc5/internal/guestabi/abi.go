package guestabi

const (
	PrecompileBase = uintptr(0x80efe000)
	PrecompileSize = uintptr(1 << 12)
	PrecompileEnd  = PrecompileBase + PrecompileSize

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

	PrecompileMagic          = uint32(0x50435250) // "PRCP"
	PrecompileVersion        = uint32(1)
	PrecompileSyscall        = uintptr(500)
	PrecompileOpcodeCompute  = uint32(1)
	PrecompileStatusReady    = uint32(0)
	PrecompileStatusSuccess  = uint32(1)
	PrecompileStatusBadInput = uint32(2)

	QEMUTestPass = uint32(0x5555)
	QEMUTestFail = uint32(0x3333)

	HeaderSize             = uintptr(24)
	MaxWords               = int((InputSize - HeaderSize) / 8)
	PrecompileHeaderSize   = uintptr(32)
	PrecompileResultOffset = uintptr(24)
	PrecompileWordsOffset  = PrecompileHeaderSize
	PrecompileMaxWords     = int((PrecompileSize - PrecompileWordsOffset) / 8)
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

type PrecompileRequest struct {
	Magic     uint32
	Version   uint32
	Opcode    uint32
	Status    uint32
	WordCount uint32
	Reserved  uint32
	Result    uint64
}
