# DIV INT_MIN / -1 must return INT_MIN (not trap). Exercises the
# OVERFLOW_DIVIDEND branch in r_type.zkc:185-186.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 0x8000000000000000
    li t1, -1
    div t2, t0, t1
    li t3, 0x8000000000000000
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
