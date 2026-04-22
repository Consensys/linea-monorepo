OUTPUT_ARCH(riscv)
ENTRY(_start)

MEMORY
{
	RAM (rwx) : ORIGIN = 0x80000000, LENGTH = 16M
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
