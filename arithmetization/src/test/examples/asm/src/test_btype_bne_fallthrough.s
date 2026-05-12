# BNE must NOT branch when registers are equal. Tests FUNCT3_BNE in b_type.zkc:75.
# Pass path: fall through to ecall. Fail path: branch to the .word 0 trap.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 42
    li t1, 42
    bne t0, t1, should_not_branch
    li a7, 93
    ecall
should_not_branch:
    .word 0
