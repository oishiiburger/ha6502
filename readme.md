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
        jmp _end    ;so does this one
        nop
        nop
_other  lda $fded,y
        nop
_end    jmp _other
```

Output currently looks like this (in addition to the object file written to disk):

```
Hobbyist's Assembler for 6502 microprocessors
(c) 2020 Dr. Christopher Graham
https://github.com/oishiiburger/ha6502

Assembly Listing =========================================================
                      | 1      ;This is a comment
                      | 2      
                      | 3              org $0800
                      | 4      
0800 _start A9 00     | 5      _start  lda #$00    ;this line has a label
0802        4C 0B 08  | 6              jmp _end    ;so does this one
0805        EA        | 7              nop
0806        EA        | 8              nop
0807 _other B9 ED FD  | 9      _other  lda $fded,y
080A        EA        | 10             nop
080B _end   4C 07 08  | 11     _end    jmp _other

Object will fill from $0800 to $080E. ($000E bytes)

Symbol Table =============================================================
_end    $080B       _other  $0807       _start  $0800       

Wrote 14 bytes to ./files/out.o.
```