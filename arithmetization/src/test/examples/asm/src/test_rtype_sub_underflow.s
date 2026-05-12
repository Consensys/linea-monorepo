# SUB: 0 - 1 == -1 (0xFFFFFFFFFFFFFFFF). Exercises the u65 prepend
# trick in r_type.zkc:108-112 that prevents two's-complement underflow.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 0
    li t1, 1
    sub t2, t0, t1
    li t3, -1
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
