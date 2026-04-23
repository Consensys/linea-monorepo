MEMORY
{
	RAM (rwx) : ORIGIN = 0x80000000, LENGTH = 0x00eff000
	GUEST_STATUS (rw) : ORIGIN = 0x80eff000, LENGTH = 0x1000
	GUEST_INPUT (rw) : ORIGIN = 0x80f00000, LENGTH = 1M
}

REGION_ALIAS("FLASH_TEXT", RAM)

_stack_size = 2K;

PHDRS
{
	text PT_LOAD FLAGS(5);
	data PT_LOAD FLAGS(6);
}

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
	} > FLASH_TEXT :text

	. = ALIGN(0x1000);

	.stack0 ALIGN(0x1000) (NOLOAD) :
	{
		. = ALIGN(16);
		. += _stack_size;
		_stack_top = .;
	} > RAM :data

	.stack1 (NOLOAD) :
	{
		. = ALIGN(16);
		. += DEFINED(__num_stacks) && __num_stacks >= 2 ? _stack_size : 0;
		_stack1_top = .;
	} > RAM :data

	.stack2 (NOLOAD) :
	{
		. = ALIGN(16);
		. += DEFINED(__num_stacks) && __num_stacks >= 3 ? _stack_size : 0;
		_stack2_top = .;
	} > RAM :data

	.stack3 (NOLOAD) :
	{
		. = ALIGN(16);
		. += DEFINED(__num_stacks) && __num_stacks >= 4 ? _stack_size : 0;
		_stack3_top = .;
	} > RAM :data

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
	} > RAM :data

	.bss :
	{
		. = ALIGN(16);
		_sbss = .;
		*(.sbss .sbss.*)
		*(.bss .bss.*)
		*(COMMON)
		. = ALIGN(16);
		_ebss = .;
	} > RAM :data

	/DISCARD/ :
	{
		*(.eh_frame)
	}
}

_heap_start = ALIGN(_ebss, 16);
_heap_end = ORIGIN(RAM) + LENGTH(RAM);
_globals_start = _sdata;
_globals_end = _ebss;
__guest_input_base = ORIGIN(GUEST_INPUT);
__guest_input_end = ORIGIN(GUEST_INPUT) + LENGTH(GUEST_INPUT);
__guest_status_base = ORIGIN(GUEST_STATUS);
__guest_status_end = ORIGIN(GUEST_STATUS) + LENGTH(GUEST_STATUS);
