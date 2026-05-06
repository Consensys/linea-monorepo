ENTRY(_start)

MEMORY {
    /* program region: 128 MiB (0x00000000 -> 0x07FFFFFF) */
    PROGRAM (rx) : ORIGIN = 0x00000000, LENGTH = 0x8000000
    /* stack region: 8 MiB (0x08000000 <- 0x087fffff) */
    STACK   (rw) : ORIGIN = 0x08000000, LENGTH = 0x800000
    /* input region: 1 GiB (0x08800000 -> 0x48800000) */
    IN  (r)  : ORIGIN = 0x08800000, LENGTH = 0x40000000
}

SECTIONS {
    . = ORIGIN(PROGRAM);

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

    /* stack grows downward from the top of the stack region */
    _stack_start = ORIGIN(STACK) + LENGTH(STACK) - 1;

    /* symbol marking the start of the input region */
    _input_start = ORIGIN(IN);
}