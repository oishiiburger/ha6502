package main

// Opcode tables
// zop, imm, zp, zpx, abs, absx, absy, zpxi, zpiy, ind, rel

var pseudoOps = map[string]string{
	"org": "Set start address for program"}

var mnemonics = map[string]string{
	"adc": "Add memory to accumulator with carry",
	"bcc": "Branch on carry clear",
	"brk": "Force break",
	"jmp": "Jump to new location",
	"lda": "Load accumulator with memory",
	"nop": "No operation",
	"pha": "Push accumulator on stack"}

var opZop = map[string]byte{
	"brk": 0x00,
	"nop": 0xea,
	"pha": 0x48}

var opImm = map[string]byte{
	"adc": 0x69,
	"lda": 0xa9}

var opZp = map[string]byte{
	"adc": 0x65,
	"lda": 0xa5}

var opZpx = map[string]byte{
	"adc": 0x75,
	"lda": 0xb5}

var opAbs = map[string]byte{
	"adc": 0x6d,
	"jmp": 0x4c,
	"lda": 0xad}

var opAbsx = map[string]byte{
	"adc": 0x7d,
	"lda": 0xbd}

var opAbsy = map[string]byte{
	"adc": 0x79,
	"lda": 0xb9}

var opZpxi = map[string]byte{
	"adc": 0x61,
	"lda": 0xa1}

var opZpiy = map[string]byte{
	"adc": 0x71,
	"lda": 0xb1}

var opInd = map[string]byte{
	"jmp": 0x6c}

var opRel = map[string]byte{
	"bcc": 0x90}
