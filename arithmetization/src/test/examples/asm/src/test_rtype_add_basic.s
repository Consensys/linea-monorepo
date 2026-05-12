# ADD: 2 + 3 == 5. Sanity check for the OP / FUNCT3_ADD branch in r_type.zkc.
# Pass path: ecall with a0=0. Fail path: execute illegal opcode → zkc exec aborts.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    li t0, 2
    li t1, 3
    add t2, t0, t1
    li t3, 5
    xor a0, t2, t3        # 0 iff pass
    bnez a0, fail
    li a7, 93
    ecall
fail:
    .word 0                # opcode 0 → UNDEFINED_TYPE → interpreter fail
