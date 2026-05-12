# REM: -7 % 3 == -1. Result must take dividend's sign. Exercises the
# negative_of_double_word path in r_type.zkc:200-211.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -7
    li t1, 3
    rem t2, t0, t1
    li t3, -1
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
