# SD + LD roundtrip on the stack. Cross-checks S-type SD (s_type.zkc:72)
# and I-type LD (i_type.zkc:93) together — storing a doubleword and reading it
# back must return the same value.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 0xCAFEBABEDEADBEEF
    addi sp, sp, -8
    sd t0, 0(sp)
    ld t1, 0(sp)
    addi sp, sp, 8
    xor a0, t0, t1
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
