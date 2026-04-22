MEMORY
{
	RAM (rwx) : ORIGIN = 0x80000000, LENGTH = 16M
}

REGION_ALIAS("FLASH_TEXT", RAM)

_stack_size = 16K;

SECTIONS
{
	.text :
	{
		. = ALIGN(4);
		KEEP(*(.init))
		*(.text.handleInterruptASM)
		*(.text .text.*)
		*(.rodata .rodata.*)
		. = ALIGN(4);
	} > FLASH_TEXT

	.stack0 (NOLOAD) :
	{
		. = ALIGN(16);
		. += _stack_size;
		_stack_top = .;
	} > RAM

	.stack1 (NOLOAD) :
	{
		. = ALIGN(16);
		. += DEFINED(__num_stacks) && __num_stacks >= 2 ? _stack_size : 0;
		_stack1_top = .;
	} > RAM

	.stack2 (NOLOAD) :
	{
		. = ALIGN(16);
		. += DEFINED(__num_stacks) && __num_stacks >= 3 ? _stack_size : 0;
		_stack2_top = .;
	} > RAM

	.stack3 (NOLOAD) :
	{
		. = ALIGN(16);
		. += DEFINED(__num_stacks) && __num_stacks >= 4 ? _stack_size : 0;
		_stack3_top = .;
	} > RAM

	_sidata = LOADADDR(.data);

	.data :
	{
		. = ALIGN(16);
		PROVIDE(__global_pointer$ = . + 0x800);
		_sdata = .;
		*(.sdata .sdata.*)
		*(.data .data.*)
		. = ALIGN(16);
		_edata = .;
	} > RAM AT>FLASH_TEXT

	.bss :
	{
		. = ALIGN(16);
		_sbss = .;
		*(.sbss .sbss.*)
		*(.bss .bss.*)
		*(COMMON)
		. = ALIGN(16);
		_ebss = .;
	} > RAM

	/DISCARD/ :
	{
		*(.eh_frame)
	}
}

_heap_start = ALIGN(_ebss, 16);
_heap_end = ORIGIN(RAM) + LENGTH(RAM);
_globals_start = _sdata;
_globals_end = _ebss;
