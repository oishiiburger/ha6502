/* 	Hobbyist's Assembler for 6502 microprocessors
	A simple assembler for little projects and tinkering
	See README.md for more information

	-> opcodes.go

=============================================================================
MIT License

Copyright (c) 2020 Dr. Christopher Graham

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
==============================================================================
*/

package main

import "regexp"

// Opcode tables
// zop, imm, zp, zpx, abs, absx, absy, zpxi, zpiy, ind, rel

// Regexp for matching operands
var rOperand = regexp.MustCompile(`[(]?[#$]*[0-9a-f]{2,4}[,xy)]*`)
var rImm = regexp.MustCompile(`^#[$]?[0-9a-f]{2}$`)
var rZp = regexp.MustCompile(`^[$]?[0-9a-f]{2}$`)
var rZpx = regexp.MustCompile(`^[$]?[0-9a-f]{2},x$`)
var rZpy = regexp.MustCompile(`^[$]?[0-9a-f]{2},y$`)
var rAbs = regexp.MustCompile(`^[$]?[0-9a-f]{4}$`)
var rAbsx = regexp.MustCompile(`^[$]?[0-9a-f]{4},x$`)
var rAbsy = regexp.MustCompile(`^[$]?[0-9a-f]{4},y$`)
var rZpxi = regexp.MustCompile(`^[(][$]?[0-9a-f]{2},x[])]$`)
var rZpiy = regexp.MustCompile(`^[(][$]?[0-9a-f]{2}[)],y$`)
var rInd = regexp.MustCompile(`^[(][$]?[0-9a-f]{2,4}[)]$`)

var rAddr = regexp.MustCompile(`[0-9a-f]{2,4}`)
var rMnem = regexp.MustCompile(`^[A-Za-z]{3}$`)

var rLabel = regexp.MustCompile(`^[A-Za-z]{1,6}$`)
var rLabelCol = regexp.MustCompile(`^[A-Za-z]{1,6}:$`)
var rLabelOpAbs = regexp.MustCompile(`^[A-Za-z]{1,6}$`)
var rLabelOpInd = regexp.MustCompile(`^[(][A-Za-z]{1,6}[)]$`)

var pseudoOps = map[string]string{
	"dfb": "Define bytes of data",         // not yet implemented
	"equ": "Store an address in a symbol", // not yet implemented
	"org": "Set start address for program"}

// including descriptions for a potential educational feature
var mnemonics = map[string]string{
	"adc": "Add memory to accumulator with carry",
	"bcc": "Branch on carry clear",
	"bcs": "Branch on carry set",
	"beq": "Branch on result zero",
	"bit": "Test bits in memory with accumulator",
	"bmi": "Branch on result minus",
	"bne": "Branch on result not zero",
	"bpl": "Branch on result plus",
	"brk": "Force break",
	"bvc": "Branch on overflow clear",
	"bvs": "Branch on overflow set",
	"clc": "Clear carry flag",
	"cld": "Clear decimal mode",
	"cli": "Clear interrupt disable status",
	"clv": "Clear overflow flag",
	"cmp": "Compare memory and accumulator",
	"cpx": "Compare memory and index X",
	"cpy": "Compare memory and index Y",
	"dec": "Decrement memory by one",
	"dex": "Decrement index X by one",
	"dey": "Decrement index Y by one",
	"eor": "'Exclusive-Or' memory with accumulator",
	"inc": "Increment memory by one",
	"inx": "Increment index X by one",
	"iny": "Increment index Y by one",
	"jmp": "Jump to new location",
	"jsr": "Jump to new location saving return address",
	"lda": "Load accumulator with memory",
	"ldx": "Load index X with memory",
	"ldy": "Load index Y with memory",
	"lsr": "Shift right one bit (memory or accumulator)",
	"nop": "No operation",
	"ora": "'OR' memory with accumulator",
	"pha": "Push accumulator on stack",
	"php": "Push processor status on stack",
	"pla": "Pull accumulator from stack",
	"plp": "Pull processor status from stack",
	"rol": "Rotate one bit left (memory or accumulator)",
	"ror": "Rotate one bit right (memory or accumulator)",
	"rti": "Return from interrupt",
	"rts": "Return from subroutine",
	"sbc": "Subtract memory from accumulator with borrow",
	"sec": "Set carry flag",
	"sed": "Set decimal mode",
	"sei": "Set interrupt disable status",
	"sta": "Store accumulator in memory",
	"stx": "Store index X in memory",
	"sty": "Store index Y in memory",
	"tax": "Transfer accumulator to index X",
	"tay": "Transfer accumulator to index Y",
	"tsx": "Transfer stack pointer to index X",
	"txa": "Transfer index X to accumulator",
	"txs": "Transfer index X to stack pointer",
	"tya": "Transfer index Y to accumulator"}

var opZop = map[string]byte{
	"brk": 0x00,
	"clc": 0x18,
	"cld": 0xd8,
	"cli": 0x58,
	"clv": 0xb8,
	"dex": 0xca,
	"dey": 0x88,
	"inx": 0xe8,
	"iny": 0xc8,
	"lsr": 0x4a,
	"nop": 0xea,
	"pha": 0x48,
	"php": 0x08,
	"pla": 0x68,
	"plp": 0x28,
	"rol": 0x2a,
	"ror": 0x6a,
	"rti": 0x40,
	"rts": 0x60,
	"sec": 0x38,
	"sed": 0xf8,
	"sei": 0x78,
	"tax": 0xaa,
	"tay": 0xa8,
	"tsx": 0xba,
	"txa": 0x8a,
	"txs": 0x9a,
	"tya": 0x98}

var opImm = map[string]byte{
	"adc": 0x69,
	"cmp": 0xc9,
	"cpx": 0xe0,
	"cpy": 0xc0,
	"eor": 0x49,
	"lda": 0xa9,
	"ldx": 0xa2,
	"ldy": 0xa0,
	"ora": 0x09,
	"sbc": 0xe9}

var opZp = map[string]byte{
	"adc": 0x65,
	"bit": 0x24,
	"cmp": 0xc5,
	"cpx": 0xe4,
	"cpy": 0xc4,
	"dec": 0xc6,
	"eor": 0x45,
	"inc": 0xe6,
	"lda": 0xa5,
	"ldx": 0xa6,
	"ldy": 0xa4,
	"lsr": 0x46,
	"ora": 0x05,
	"rol": 0x26,
	"ror": 0x66,
	"sbc": 0xe5,
	"sta": 0x85,
	"stx": 0x86,
	"sty": 0x84}

var opZpx = map[string]byte{
	"adc": 0x75,
	"cmp": 0xd5,
	"dec": 0xd6,
	"eor": 0x55,
	"inc": 0xf6,
	"lda": 0xb5,
	"ldy": 0xb4,
	"lsr": 0x56,
	"ora": 0x15,
	"rol": 0x36,
	"ror": 0x76,
	"sbc": 0xf5,
	"sta": 0x95,
	"stx": 0x96,
	"sty": 0x94}

var opZpy = map[string]byte{
	"ldx": 0xb6}

var opAbs = map[string]byte{
	"adc": 0x6d,
	"bit": 0x2c,
	"cmp": 0xcd,
	"cpx": 0xec,
	"cpy": 0xcc,
	"dec": 0xce,
	"eor": 0x4d,
	"inc": 0xee,
	"jmp": 0x4c,
	"jsr": 0x20,
	"lda": 0xad,
	"ldx": 0xae,
	"ldy": 0xac,
	"lsr": 0x4e,
	"ora": 0x0d,
	"rol": 0x2e,
	"ror": 0x6e,
	"sbc": 0xed,
	"sta": 0x8d,
	"stx": 0x8e,
	"sty": 0x8c}

var opAbsx = map[string]byte{
	"adc": 0x7d,
	"cmp": 0xdd,
	"dec": 0xde,
	"eor": 0x5d,
	"inc": 0xfe,
	"lda": 0xbd,
	"ldy": 0xbc,
	"lsr": 0x5e,
	"ora": 0x1d,
	"rol": 0x3e,
	"ror": 0x7e,
	"sbc": 0xfd,
	"sta": 0x9d}

var opAbsy = map[string]byte{
	"adc": 0x79,
	"cmp": 0xd9,
	"eor": 0x59,
	"lda": 0xb9,
	"ldx": 0xbe,
	"ora": 0x19,
	"sbc": 0xf9,
	"sta": 0x99}

var opZpxi = map[string]byte{
	"adc": 0x61,
	"cmp": 0xc1,
	"eor": 0x41,
	"lda": 0xa1,
	"ora": 0x01,
	"sbc": 0xe1,
	"sta": 0x81}

var opZpiy = map[string]byte{
	"adc": 0x71,
	"cmp": 0xd1,
	"eor": 0x51,
	"lda": 0xb1,
	"ora": 0x11,
	"sbc": 0xf1,
	"sta": 0x91}

var opInd = map[string]byte{
	"jmp": 0x6c}

var opRel = map[string]byte{
	"bcc": 0x90,
	"bcs": 0xb0,
	"beq": 0xf0,
	"bmi": 0x30,
	"bne": 0xd0,
	"bpl": 0x10,
	"bvc": 0x50,
	"bvs": 0x70}
