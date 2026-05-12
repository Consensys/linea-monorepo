# SLTI signed: -1 < 1 should set rd to 1. Tests FUNCT3_SLTI in i_type.zkc:113-120.
# SLTI sign-extends the 12-bit immediate; comparison is signed (so -1 < 1 is true).
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -1
    slti t1, t0, 1
    li t2, 1
    xor a0, t1, t2
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
