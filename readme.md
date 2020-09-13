ha6502 -- Hobbyist's Assembler for 6502 Microprocessors
=======================================================

This is a simple 2-pass assembler I am writing for fun. Features are still being added and the instruction set is not yet complete.

## Features
* Labels for automated addressing
* A few pseudo-ops (org, equ, dfb)
* Pretty printing of the object code next to the listing
* Symbol table

Source files should feature a format similar to the following. Note that addresses must be either 2 (zero-page) or 4 (elsewhere) hex digits long, i.e. you must have leading 0s:

```
;This is a comment

        org $0800

start:  lda #$00    ;this line has a label
        nop
        lda ($ff,x)
        nop
        nop
end:    jmp start   ;so does this one
```