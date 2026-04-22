OUTPUT_ARCH(riscv)
ENTRY(_start)

MEMORY
{
	RAM (rwx) : ORIGIN = 0x80000000, LENGTH = 0x00eff000
	GUEST_STATUS (rw) : ORIGIN = 0x80eff000, LENGTH = 0x1000
	GUEST_INPUT (rw) : ORIGIN = 0x80f00000, LENGTH = 1M
}

SECTIONS
{
	. = ORIGIN(RAM);

	.text :
	{
		KEEP(*(.text.init))
		*(.text .text.*)
		*(.rodata .rodata.*)
		. = ALIGN(4);
	} > RAM

	.data ALIGN(0x1000) :
	{
		PROVIDE(__global_pointer$ = . + 0x800);
		*(.sdata .sdata.*)
		*(.data .data.*)
	} > RAM

	. = ALIGN(16);
	.bss (NOLOAD) :
	{
		__bss_start = .;
		*(.sbss .sbss.*)
		*(.bss .bss.*)
		*(COMMON)
		__bss_end = .;
	} > RAM

	. = ALIGN(16);
	.stack (NOLOAD) :
	{
		. += 64K;
		__stack_top = .;
	} > RAM

	/DISCARD/ :
	{
		*(.comment)
		*(.eh_frame*)
		*(.note*)
	}
}

__guest_input_base = ORIGIN(GUEST_INPUT);
__guest_input_end = ORIGIN(GUEST_INPUT) + LENGTH(GUEST_INPUT);
__guest_status_base = ORIGIN(GUEST_STATUS);
__guest_status_end = ORIGIN(GUEST_STATUS) + LENGTH(GUEST_STATUS);
