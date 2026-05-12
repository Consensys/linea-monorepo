# JAL unconditional forward jump. If JAL works, control jumps over the .word 0
# trap to `target` and exits cleanly. If JAL is broken (no jump or wrong offset),
# execution falls into the trap. Tests JAL in j_type.zkc.
#
# Note: uses t0 (not x0) as rd. Idiomatically `jal x0, target` discards the return
# address since x0 is hardwired to 0 in RISC-V, but zkC's register file does not
# special-case x0 — writing PC+4 to x0 corrupts subsequent reads of x0 (e.g.
# `addi a7, x0, 93` in `li a7, 93` would give a7 = 12 + 93 = 105, not 93).
# Using t0 sidesteps the issue; the JAL semantics being tested are unaffected.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    jal t0, target                  # t0 holds the (unused) return address
    .word 0                         # should be skipped
target:
    li a7, 93
    ecall
