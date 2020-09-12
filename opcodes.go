package main

// Opcode tables
// zop, imm, zp, zpx, abs, absx, absy, zpxi, zpiy, ind, rel

var mnemonics = map[string]string{
	"brk": "Force break",
	"lda": "Load accumulator with memory",
	"nop": "No operation",
	"pha": "Push accumulator on stack"}

var opImm = map[string]byte{
	"brk": 0x00,
	"lda": 0xa9,
	"nop": 0xea,
	"pha": 0x48}
