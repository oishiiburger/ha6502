Hobbyist's Assembler for 6502 Microprocessors
=============================================

This is a simple 2-pass assembler I am writing for fun. Features are still being added and the instruction set is not yet complete.

## Features
* Labels for automated addressing (of the form `_label`)
* A few pseudo-ops (org, equ, dfb)
* Pretty printing of the object code next to the listing
* Symbol table

Source files should feature a format similar to the following. Note that addresses must be either 2 (zero-page) or 4 (elsewhere) hex digits long, i.e. you must have leading 0s:

```
;This is a comment

        org $0800

_start  lda #$00    ;this line has a label
        nop
        lda ($ff,x)
        nop
        nop
_end    jmp _start   ;so does this one
_more   pha
        lda $4ab2,y
_label  nop
```

Output currently looks like this (in addition to the object file written to disk):

```
Hobbyist Assembler for 6502 microprocessors

Assembly Listing =========================================================
                      | 1      ;This is a comment
                      | 2      
                      | 3              org $0800
                      | 4      
0800 _start A9 00     | 5      _start  lda #$00    ;this line has a label
0802        EA        | 6              nop
0803        A1 FF     | 7              lda ($ff,x)
0805        EA        | 8              nop
0806        EA        | 9              nop
0807 _end   4C 00 08  | 10     _end    jmp _start   ;so does this one
080A _more  48        | 11     _more   pha
080B        B9 B2 4A  | 12             lda $4ab2,y
080E _label EA        | 13     _label  nop

Object will fill from $0800 to $080F. ($000F bytes)

Symbol Table =============================================================
_end    $0807       _label  $080E       _more   $080A       _start  $0800       

Wrote 15 bytes to ./files/out.o.
```