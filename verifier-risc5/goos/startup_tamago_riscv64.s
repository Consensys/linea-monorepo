//go:build tamago && riscv64 && tamago_libriscv

#include "textflag.h"

TEXT ·CPUInit(SB),NOSPLIT|NOFRAME,$0
	MOV	$0x80000000, T0
	MOV	$0x01200000, T1
	MOV	$0x00001000, T2
	ADD	T1, T0
	SUB	T2, T0
	MOV	T0, X2
	JMP	runtime·rt0_riscv64_tamago(SB)

TEXT ·Hwinit0(SB),NOSPLIT|NOFRAME,$0
	RET

TEXT ·Printk(SB),NOSPLIT|NOFRAME,$0-8
	RET

TEXT ·writeFinisher(SB),NOSPLIT|NOFRAME,$0-4
	MOVWU	value+0(FP), T0
	MOV	$0x00100000, T1
	MOVW	T0, 0(T1)
	RET
