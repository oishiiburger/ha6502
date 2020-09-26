Hobbyist's Assembler for 6502 microprocessors
=============================================

This is a simple 2-pass assembler. Features are still being added. It's not efficient, but it's fun to tinker with.

## Features
* Labels for automated addressing
* A few pseudo-ops (only org and equ are currently implemented)
* Pretty printing of the object code next to the listing
* Symbol table

Source files should use a format similar to the following. Note that addresses must be either 2 (zero-page) or 4 (elsewhere) hex digits long, i.e. you must have leading 0s:

```
; test program
; will ring the bell on an Apple II
; also does some useless stuff with the x register

        org $5000
bell    equ $fbe4       ;subroutine in ROM

start:  ldx #$00        ;x = 0
        cpx #$ff
        beq ring        ;ring bell if x == $ff
        inx             ;otherwise increment
        jmp start
ring:   jsr bell
        rts
        brk
```

Output currently looks like this (in addition to the object file written to disk):

```
Hobbyist's Assembler for 6502 microprocessors
https://github.com/oishiiburger/ha6502

Assembly Listing =========================================================
                      | 1      ; test program
                      | 2      ; will ring the bell on an Apple II
                      | 3      ; also does some useless stuff with the x register
                      | 4      
                      | 5              org $5000
5000        00        | 6      bell    equ $fbe4       ;subroutine in ROM
                      | 7      
5001        A2 00     | 8      start:  ldx #$00        ;x = 0
5003        E0 FF     | 9              cpx #$ff
5005        F0 03     | 10             beq ring        ;ring bell if x == $ff
5007        E8        | 11             inx             ;otherwise increment
5008        4C 00 50  | 12             jmp start
500B        20 E4 FB  | 13     ring:   jsr bell
500E        60        | 14             rts
500F        00        | 15             brk

Object will fill from $5000 through $500F. ($0010 bytes)

Symbol Table =============================================================
bell    $FBE4       ring    $500A       start   $5000       

Wrote 16 bytes to ./files/out.o.
```