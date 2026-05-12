# SB must write exactly one byte. Pre-fill 8 bytes with 0xFF, store 0x42 as a
# byte at offset 0, then read back the full doubleword: expect 0xFFFFFFFFFFFFFF42
# (little-endian: byte 0 is the lowest byte).
# Tests FUNCT3_SB in s_type.zkc:60.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    addi sp, sp, -8
    li t0, -1                       # 0xFFFFFFFFFFFFFFFF
    sd t0, 0(sp)                    # prime memory with all 1s
    li t1, 0x42
    sb t1, 0(sp)                    # only low byte must change
    ld t2, 0(sp)
    addi sp, sp, 8
    li t3, 0xFFFFFFFFFFFFFF42
    xor a0, t2, t3
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0
