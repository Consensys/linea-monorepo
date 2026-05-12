# SRA: -8 >> 1 == -4. Exercises the bit_flip-twice sign-propagation
# path in r_type.zkc:134-146. Uses R-type SRA (variable shift), not SRAI.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -8
    li t1, 1
    sra t2, t0, t1
    li t3, -4
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
