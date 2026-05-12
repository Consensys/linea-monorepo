# SRAI on negative: -16 >> 2 == -4. Tests FUNCT3_SRAI+FUNCT6_SRAI sign-propagation
# branch in i_type.zkc:144.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -16
    srai t1, t0, 2
    li t2, -4
    xor a0, t1, t2
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
