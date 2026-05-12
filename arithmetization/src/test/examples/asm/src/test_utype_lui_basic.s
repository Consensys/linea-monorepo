# LUI loads a 20-bit immediate into bits [31:12], zero-filling [11:0] and
# sign-extending to 64 bits. lui rd, 0x12345 → rd = 0x12345000.
# Tests LUI case in u_type.zkc:48.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    lui t0, 0x12345
    li t1, 0x12345000
    xor a0, t0, t1
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
