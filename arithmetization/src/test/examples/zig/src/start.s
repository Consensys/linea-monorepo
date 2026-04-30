# start.s
.section .text
.global _start
_start:
    li sp, 0x7FFFFFF  # set stack pointer to a known memory region
    call main