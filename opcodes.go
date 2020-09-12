package main

// Opcode tables
// zop, imm, zp, zpx, abs, absx, absy, zpxi, zpiy, ind, rel

var pseudoOps = map[string]string{
	"org": "Set start address for program"}

var mnemonics = map[string]string{
	"brk": "Force break",
	"lda": "Load accumulator with memory",
	"nop": "No operation",
	"pha": "Push accumulator on stack"}

var opZop = map[string]byte{
	"brk": 0x00,
	"nop": 0xea,
	"pha": 0x48}

var opImm = map[string]byte{
	"lda": 0xa9}
