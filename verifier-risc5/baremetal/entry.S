.section .text.init, "ax", @progbits
.globl _start

_start:
	.option push
	.option norelax
	la gp, __global_pointer$
	.option pop

	la sp, __stack_top
	la t0, __bss_start
	la t1, __bss_end

1:
	bgeu t0, t1, 2f
	sd zero, 0(t0)
	addi t0, t0, 8
	j 1b

2:
	call guest_main

3:
	wfi
	j 3b
