# ADDI: 5 + (-3) == 2. Sanity check for OP_IMM/FUNCT3_ADDI in i_type.zkc:110-112.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 5
    addi t1, t0, -3
    li t2, 2
    xor a0, t1, t2
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
