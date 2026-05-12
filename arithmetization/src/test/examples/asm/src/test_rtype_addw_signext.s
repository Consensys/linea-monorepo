# ADDW: 0x7FFFFFFF + 1 == 0x80000000 in 32-bit, sign-extended to
# 0xFFFFFFFF80000000 in 64-bit. Exercises sgn_extension_u32_u64
# in r_type.zkc:238-241.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 0x7FFFFFFF
    li t1, 1
    addw t2, t0, t1
    li t3, 0xFFFFFFFF80000000
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
