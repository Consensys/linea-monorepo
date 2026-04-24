.section .text
.globl _start
.type _start, @function

_start:
	/* Preserve a0/a1 as provided by QEMU and jump straight to the ELF entry. */
	li t0, ENTRY_ADDR
	jr t0
