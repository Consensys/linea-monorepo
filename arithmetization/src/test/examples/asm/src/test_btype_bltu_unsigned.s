# BLTU (unsigned) must branch: 1 < 0xFFFFFFFFFFFFFFFF as unsigned.
# Same operands as test_btype_blt_signed but with reversed roles — under signed
# comparison -1 < 1 (BLT taken with t0=-1,t1=1); under unsigned -1 is 0xFFF...
# which is the max, so 1 < -1 (BLTU taken with t0=1,t1=-1).
# Tests FUNCT3_BLTU in b_type.zkc:98.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 1
    li t1, -1
    bltu t0, t1, taken
    .word 0
taken:
    li a7, 93
    ecall
