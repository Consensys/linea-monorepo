# MULH(-2, 3) == -1: high 64 bits of (-2 * 3) = -6 = 0xFFFF…FFFA.
# Exercises the ~x+1 two's-complement step in r_type.zkc:162-169.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -2
    li t1, 3
    mulh t2, t0, t1
    li t3, -1
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
