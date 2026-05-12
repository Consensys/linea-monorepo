# DIV by zero must return -1 (RISC-V spec). Exercises the early-return
# branch in r_type.zkc:181-184.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 5
    li t1, 0
    div t2, t0, t1
    li t3, -1
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
