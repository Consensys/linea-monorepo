# BEQ taken when registers equal. Tests FUNCT3_BEQ in b_type.zkc:68.
# Pass path: branch over the .word 0 trap to `taken`. Fail path: fall through to trap.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 42
    li t1, 42
    beq t0, t1, taken
    .word 0                         # should be jumped over
taken:
    li a7, 93
    ecall
