# AUIPC with imm20=0 puts the address of the AUIPC instruction itself into rd.
# Two back-to-back AUIPCs must therefore differ by exactly 4 bytes (one instruction).
# Tests AUIPC case in u_type.zkc:44.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    auipc t0, 0                     # t0 = PC of this auipc
    auipc t1, 0                     # t1 = PC of the next auipc = t0 + 4
    sub t2, t1, t0
    li t3, 4
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
