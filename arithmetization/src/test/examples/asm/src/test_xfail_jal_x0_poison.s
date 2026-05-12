# XFAIL — JAL with rd=x0, the canonical RISC-V "unconditional forward jump"
# idiom (also what the `j target` pseudo-instruction expands to). This pattern
# currently traps in zkC and is *expected* to fail until x0 hardwiring lands.
#
# The bug, end-to-end:
#   - `j_type.zkc:47`: JAL does `registers[rd] = pc + 4` unconditionally.
#   - With rd = x0, that should be a no-op (x0 is hardwired to 0 in RISC-V),
#     but zkC's register file does not special-case x0 — so register 0 ends up
#     holding pc + 4 = 12 (= the address of the .word 0 trap below).
#   - The subsequent `li a7, 93` expands to `addi a7, x0, 93`. The ADDI reads
#     register 0 expecting 0 and gets 12, so a7 = 12 + 93 = 105.
#   - zkc exec sees syscall 105 instead of 93 (exit), panics on
#     "unsupported system call", and the process exits non-zero.
#
# When the planned fix lands (TODO at `interpreter.zkc:23-24` — the team will
# implement it in the registers memory primitive itself), this test will exit
# cleanly via ecall. At that point the test-xfail target will report
# [UNEXPECTED PASS] and this file should be either deleted or renamed without
# the test_xfail_ prefix and moved into the positive J-type suite as a
# regression guard for x0 hardwiring.
.section .text
.global _start
_start:
    li sp, 0x087fffff
    jal x0, target                  # canonical RISC-V "discard return address"
    .word 0                         # should be skipped when JAL+x0 works
target:
    li a7, 93
    ecall
