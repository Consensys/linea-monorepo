# BLT (signed) must branch: -1 < 1 as signed integers. Tests FUNCT3_BLT in b_type.zkc:82.
# Companion to test_btype_bltu_unsigned which uses the same operands with BLTU
# and must take the opposite path.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -1
    li t1, 1
    blt t0, t1, taken
    .word 0
taken:
    li a7, 93
    ecall
