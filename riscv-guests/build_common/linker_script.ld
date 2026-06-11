/* Linea zkVM memory layout — kept in sync with arithmetization/src/test/examples (PR #3332).
 *   STACK   0x00000000   8 MiB    grows down from _stack_start (= SP = 0x00800000) toward _stack_end
 *   PROGRAM 0x00800000  128 MiB   code + rodata + data + bss
 *   IN      0x08800000    1 GiB   memory-mapped input
 *   HEAP    0x48800000      ∞     grows up from _heap_start (a symbol, not a bounded region)
 */
ENTRY(_start)

MEMORY {
    STACK   (rw) : ORIGIN = 0x00000000, LENGTH = 0x00800000
    PROGRAM (rx) : ORIGIN = 0x00800000, LENGTH = 0x08000000
    IN      (r)  : ORIGIN = 0x08800000, LENGTH = 0x40000000
    /* HEAP (rw) : ORIGIN = 0x48800000, LENGTH = infinite — exposed as _heap_start below, not a region */
}

SECTIONS {
    . = ORIGIN(PROGRAM);
    _program_start = ORIGIN(PROGRAM);

    .text : {
        *(.text .text.*)
    } > PROGRAM

    .rodata : {
        *(.rodata .rodata.*)
    } > PROGRAM

    .data : {
        *(.data .data.*)
    } > PROGRAM

    .bss : {
        *(.bss .bss.*)
    } > PROGRAM

    /* stack grows downward from the top of the STACK region */
    _stack_end = ORIGIN(STACK);
    _stack_start = ORIGIN(STACK) + LENGTH(STACK);

    /* program + input region bounds */
    _program_end = ORIGIN(PROGRAM) + LENGTH(PROGRAM);
    _in_start = ORIGIN(IN);
    _in_end = ORIGIN(IN) + LENGTH(IN);

    /* heap grows upward from here by program convention (no bounded region). */
    _heap_start = ORIGIN(IN) + LENGTH(IN);
}
