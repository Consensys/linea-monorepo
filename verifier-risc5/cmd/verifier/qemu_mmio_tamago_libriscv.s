//go:build baremetal && tamago_libriscv

#include "textflag.h"

TEXT ·writeQEMUTest(SB),NOSPLIT|NOFRAME,$0-4
	MOVWU	value+0(FP), T0
	MOV	$0x00100000, T1
	MOVW	T0, 0(T1)
	RET
