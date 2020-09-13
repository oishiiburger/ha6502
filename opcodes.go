package main

import "regexp"

// Opcode tables
// zop, imm, zp, zpx, abs, absx, absy, zpxi, zpiy, ind, rel

// Regexp for matching operands
var rOperand = regexp.MustCompile(`[(]?[#$]*[0-9a-f]{2,4}[,xy)]*`)
var rImm = regexp.MustCompile(`^#[$]?[0-9a-f]{2}$`)
var rZp = regexp.MustCompile(`^[$]?[0-9a-f]{2}$`)
var rZpx = regexp.MustCompile(`^[$]?[0-9a-f]{2},x$`)
var rAbs = regexp.MustCompile(`^[$]?[0-9a-f]{4}$`)
var rAbsx = regexp.MustCompile(`^[$]?[0-9a-f]{4},x$`)
var rAbsy = regexp.MustCompile(`^[$]?[0-9a-f]{4},y$`)
var rZpxi = regexp.MustCompile(`^[(][$]?[0-9a-f]{2},x[])]$`)
var rZpiy = regexp.MustCompile(`^[(][$]?[0-9a-f]{2}[)],y$`)
var rInd = regexp.MustCompile(`^[(][$]?[0-9a-f]{2,4}[)]$`)

var rAddr = regexp.MustCompile(`[0-9a-f]{2,4}`)
var rMnem = regexp.MustCompile(`^[A-Za-z]{3}$`)

var rLabel = regexp.MustCompile(`^[A-Za-z]{1,6}[:]$`)
var rLabelSolo = regexp.MustCompile(`^[A-Za-z]{1,6}$`)
var rLabelOpAbs = regexp.MustCompile(`^[A-Za-z]{1,6}$`)
var rLabelOpInd = regexp.MustCompile(`^[(][A-Za-z]{1,6}[)]$`)

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
