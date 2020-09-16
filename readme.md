Hobbyist's Assembler for 6502 microprocessors
=============================================

This is a simple 2-pass assembler. Features are still being added. It's not efficient, but it's fun to tinker with.

## Features
* Labels for automated addressing
* A few pseudo-ops (only org is currently implemented)
* Pretty printing of the object code next to the listing
* Symbol table

Source files should use a format similar to the following. Note that addresses must be either 2 (zero-page) or 4 (elsewhere) hex digits long, i.e. you must have leading 0s:

```
;Test program
        org $5000

        brk
start:  ldx #$0f
        cpx #$00
        beq end
        dex
end:    bne start
        brk
timing: nop
        nop
        nop
        jmp end         ;back to end
        brk
```

Output currently looks like this (in addition to the object file written to disk):

```
Hobbyist's Assembler for 6502 microprocessors
https://github.com/oishiiburger/ha6502

Assembly Listing =========================================================
                      | 1      ;Test program
                      | 2              org $5000
                      | 3      
5000        00        | 4              brk
5001 start  A2 0F     | 5      start:  ldx #$0f
5003        E0 00     | 6              cpx #$00
5005        F0 01     | 7              beq end
5007        CA        | 8              dex
5008 end    D0 F7     | 9      end:    bne start
500A        00        | 10             brk
500B timing EA        | 11     timing: nop
500C        EA        | 12             nop
500D        EA        | 13             nop
500E        4C 08 50  | 14             jmp end         ;back to end
5011        00        | 15             brk

Object will fill from $5000 through $5011. ($0012 bytes)

Symbol Table =============================================================
end     $5008       start   $5001       timing  $500B       

Wrote 18 bytes to ./files/out.o.
```