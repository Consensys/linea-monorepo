# SRAIW: word-shift on 32-bit value, sign-extended to 64-bit.
# -16 (as i32) >> 2 == -4; result must be sign-extended to i64 (-4 as 0xFFFFFFFFFFFFFFFC).
# Tests SRAIW in OP_IMM_32 branch of i_type.zkc (RV64 *W variant).
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, -16
    sraiw t1, t0, 2
    li t2, -4
    xor a0, t1, t2
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
