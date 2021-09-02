/* 	Hobbyist's Assembler for 6502 microprocessors
A simple assembler for little projects and tinkering
See README.md for more information

-> main.go

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

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

var info = map[string]string{
	"title":      "Hobbyist's Assembler for 6502 microprocessors",
	"github":     "https://github.com/oishiiburger/ha6502",
	"shortTitle": "ha6502",
}

var continueOnError bool = false

type instruction struct {
	mnemonic   string
	kind       string
	opcode     byte
	length     int
	opLowByte  byte
	opHighByte byte
	label      string
	isComment  bool
}

type symbol struct {
	label       string
	addLowByte  byte
	addHighByte byte
	intAddr     int
}

// Consts
const comchars string = ";*"
const modchars string = "#$"
const maxLabelLength int = 7

// Globals
var filename string
var ofilename string
var curLine int = 0 // set to 0 when not testing
var org int = 0
var pass int = 1
var symbols []symbol
var lines []string

func main() {
	var pass1Inst []instruction
	var pass2Inst []instruction

	fmt.Println(info["title"] + "\n" + info["github"])

	if len(os.Args) <= 1 {
		errHandler(errs["nofile"])
	} else if len(os.Args) == 2 {
		filename = os.Args[1]
		ofilename = removePathFileExtension(filename) + ".o"
		fmt.Println(ofilename)
	} else {
		errHandler(errs["toomanyargs"])
	}

	lines = loadFile(filename)

	var objectCode [][]byte

	pass1Inst = runPass(lines, pass1Inst)

	getOrg(pass1Inst)
	getSymbols(pass1Inst)

	pass++

	pass2Inst = runPass(lines, pass2Inst)

	objectCode = asmObject(pass2Inst)

	printAssembly(lines, objectCode)
	printSymbolTable()
	saveFile(ofilename, objectCode)
}

func runPass(lines []string, insts []instruction) []instruction {
	for i, line := range lines {
		curLine = i
		insts = append(insts, parseLine(line))
	}
	return insts
}

func loadFile(filename string) (lines []string) {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		errHandler(errs["file"])
	}
	lines = strings.Split(string(file), "\n")
	return lines
}

// Parses each line. Strips comments. Checks for labels and assigns mnemonics. Handles pseudo-ops.
func parseLine(line string) (cur instruction) {
	if len(line) == 0 { // ignore blank lines
		cur.isComment = true
		return cur
	}
	if strings.ContainsAny(line, comchars) { // strip comments
		for _, char := range comchars {
			line = strings.Split(line, string(char))[0]
		}
		if len(strings.TrimSpace(line)) == 0 { // if the entire line is a comment
			cur.isComment = true
			return cur
		}
	}
	cur.isComment = false
	lineArr := strings.Fields(line)
	if len(lineArr) > 0 && len(lineArr) < 4 {
		if rLabelCol.MatchString(lineArr[0]) {
			cur.label = strings.ReplaceAll(lineArr[0], ":", "")
			if len(cur.label) > maxLabelLength {
				errHandler(errs["labelLength"])
			}
			if len(lineArr) > 1 {
				if isMnemonic(lineArr[1]) { // mnemonic, label
					cur.mnemonic = strings.ToLower(lineArr[1])
					if len(lineArr) == 2 {
						cur.kind = "zop"
						cur.length = 1
					}
				} else {
					errHandler(errs["mnemonic"])
				}
				if len(lineArr) > 2 {
					cur = parseOperand(lineArr[2], cur)
				}
			}
		} else if rLabel.MatchString(lineArr[0]) && !isMnemonic(lineArr[0]) { // Pseudo-ops
			if pass == 1 {
				if len(lineArr) > 1 && isMnemonic(lineArr[1]) { // pseudoop mnemonic, label
					cur.mnemonic = strings.ToLower(lineArr[1])
					cur.kind = "pse"
					cur.isComment = true
					if len(lineArr) != 3 {
						errHandler(errs["parser"], "Pseudo-op is missing arguments.")
					} else {
						cur = parseAddress(rAddr.FindString(lineArr[2]), cur, symbols)
						if cur.mnemonic == "equ" {
							// load the equate label into the symbol table with the operand address
							var tmp symbol
							tmp.label = lineArr[0]
							if symbolExists(tmp.label) {
								errHandler(errs["duplicatesym"])
							}
							tmp.addHighByte = cur.opHighByte
							tmp.addLowByte = cur.opLowByte
							var tmpAddr = [2]byte{tmp.addHighByte, tmp.addLowByte}
							tmp.intAddr = hexToInt(tmpAddr)
							symbols = append(symbols, tmp)
						}
					}
				} else {
					errHandler(errs["mnemonic"], "Expected a pseudo-op.")
				}
			}
		} else {
			if isMnemonic(lineArr[0]) { // mnemonic, no label
				cur.mnemonic = strings.ToLower(lineArr[0])
				if len(lineArr) == 1 {
					cur.kind = "zop"
					cur.length = 1
				}
			} else {
				errHandler(errs["mnemonic"], "(Is there an ill-formed label?)")
			}
			if len(lineArr) > 1 {
				cur = parseOperand(lineArr[1], cur)
			}
		}
	} else {
		errHandler(errs["parser"], "Too many elements in line.")
	}
	cur = assignOpcode(cur)
	return cur
}

func parseOperand(op string, inst instruction) instruction {
	op = strings.ToLower(op)
	if rOperand.MatchString(op) && !symbolExists(op) { // is not a label
		switch {
		case rImm.MatchString(op):
			inst.kind = "imm"
			inst.length = 2
		case rInd.MatchString(op):
			inst.kind = "ind"
			inst.length = 3
		case rZp.MatchString(op):
			inst.kind = "zp"
			inst.length = 2
		case rZpx.MatchString(op):
			inst.kind = "zpx"
			inst.length = 2
		case rZpy.MatchString(op):
			inst.kind = "zpy"
			inst.length = 2
		case rZpiy.MatchString(op):
			inst.kind = "zpiy"
			inst.length = 2
		case rZpxi.MatchString(op):
			inst.kind = "zpxi"
			inst.length = 2
		case rAbs.MatchString(op):
			if _, ok := opRel[inst.mnemonic]; ok { // check if actually a relative instruction (branches)
				inst.kind = "rel"
				inst.length = 2
			} else {
				inst.kind = "abs"
				inst.length = 3
			}
		case rAbsx.MatchString(op):
			inst.kind = "absx"
			inst.length = 3
		case rAbsy.MatchString(op):
			inst.kind = "absy"
			inst.length = 3
		default:
			if len(op) > 4 {
				errHandler(errs["parser"], "Operand/address is ill formed or does not match template.")
			} else if !(len(op) == 2 || len(op) == 4) {
				errHandler(errs["parser"], "Address is not 2 or 4 characters.")
			} else {
				errHandler(errs["parser"], "Does not match any operand template.")
			}
		}
		inst = parseAddress(rAddr.FindString(op), inst, symbols)
	} else { // handle labels
		switch {
		// improvements needed: check if symbols known; if so, check addresses for ZP
		// add handling of equates in indirect instructions
		case rLabelOpAbs.MatchString(op):
			if !symbolExists(op) && pass > 1 {
				errHandler(errs["unknownsym"])
			}
			if _, ok := opRel[inst.mnemonic]; ok { // check if actually a relative instruction (branches)
				inst.kind = "rel"
				inst.length = 2
			} else if strings.ToLower(op) == "a" && opZop[inst.mnemonic] > 0 { // check for ror a, rol a, similar
				inst.kind = "zop"
				inst.length = 1
			} else {
				inst.kind = "abs"
				inst.length = 3
			}
		case rLabelOpInd.MatchString(op):
			if _, ok := opRel[inst.mnemonic]; ok { // check if actually a relative instruction (branches)
				inst.kind = "rel"
				inst.length = 2
			} else {
				inst.kind = "ind"
				inst.length = 3
			}
		}
		found := rLabel.FindString(op)
		if len(found) > 0 {
			inst = parseAddress(found, inst, symbols)
		} else {
			errHandler(errs["label"])
		}
	}
	return inst
}

func parseAddress(addr string, inst instruction, symbols []symbol) instruction {
	if rAddr.MatchString(addr) && !symbolExists(addr) { // if it looks like an address and is not a known symbol
		bytes, e := hex.DecodeString(addr)
		if e != nil {
			errHandler(errs["conversion"])
		}
		if len(addr)/2 != inst.length-1 && addr[:2] != "00" && inst.kind != "rel" && inst.kind != "pse" {
			errHandler(errs["length"], "Expected "+strconv.Itoa(inst.length-1)+" bytes for "+inst.mnemonic+".")
		} else {
			if len(bytes) > 1 {
				inst.opHighByte = bytes[0]
				inst.opLowByte = bytes[1]
			} else {
				inst.opLowByte = bytes[0]
			}
		}
	} else if rLabel.MatchString(addr) {
		for _, symbol := range symbols {
			if symbol.label == addr { // symbol is known from previous pass
				bytes := intToHex(symbol.intAddr)
				if len(addr)/2 != inst.length-1 && inst.kind != "rel" && inst.kind != "pse" {
					errHandler(errs["length"], "Expected "+strconv.Itoa(inst.length-1)+" bytes for "+inst.mnemonic+".")
				} else {
					if len(bytes) > 1 {
						inst.opHighByte = bytes[0]
						inst.opLowByte = bytes[1]
					} else {
						inst.opLowByte = bytes[0]
					}
				}
			}
		}
	} else {
		errHandler(errs["conversion"])
	}
	return inst
}

func assignOpcode(inst instruction) instruction {
	if !inst.isComment && inst.mnemonic != "" {
		_, ok := pseudoOps[inst.mnemonic]
		if ok {
			inst.kind = "pse"
			inst.isComment = true // set pseudo-ops to comments
		} else {
			switch inst.kind {
			case "zop":
				_, ok = opZop[inst.mnemonic]
				if ok {
					inst.opcode = opZop[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not a zero-operand instruction.")
				}
			case "imm":
				_, ok = opImm[inst.mnemonic]
				if ok {
					inst.opcode = opImm[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an immediate instruction.")
				}
			case "zp":
				_, ok = opZp[inst.mnemonic]
				if ok {
					inst.opcode = opZp[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not a zero-page instruction.")
				}
			case "zpx":
				_, ok = opZpx[inst.mnemonic]
				if ok {
					inst.opcode = opZpx[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not a zero-page,X instruction.")
				}
			case "zpy":
				_, ok = opZpy[inst.mnemonic]
				if ok {
					inst.opcode = opZpy[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not a zero-page,Y instruction.")
				}
			case "abs":
				_, ok = opAbs[inst.mnemonic]
				if ok {
					inst.opcode = opAbs[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an absolute instruction.")
				}
			case "absx":
				_, ok = opAbsx[inst.mnemonic]
				if ok {
					inst.opcode = opAbsx[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an absolute,X instruction.")
				}
			case "absy":
				_, ok = opAbsy[inst.mnemonic]
				if ok {
					inst.opcode = opAbsy[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an absolute,y instruction.")
				}
			case "zpxi":
				_, ok = opZpxi[inst.mnemonic]
				if ok {
					inst.opcode = opZpxi[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an indexed indirect instruction.")
				}
			case "zpiy":
				_, ok = opZpiy[inst.mnemonic]
				if ok {
					inst.opcode = opZpiy[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an indirect indexed instruction.")
				}
			case "ind":
				_, ok = opInd[inst.mnemonic]
				if ok {
					inst.opcode = opInd[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not an indirect instruction.")
				}
			case "rel":
				_, ok = opRel[inst.mnemonic]
				if ok {
					inst.opcode = opRel[inst.mnemonic]
				} else {
					errHandler(errs["opcode"], "Not a relative instruction.")
				}
			default:
				errHandler(errs["opcode"])
			}
		}
	}
	return inst
}

func asmObject(insts []instruction) (obj [][]byte) {
	var PC int = org
	for i, inst := range insts {
		if PC > 0xffff {
			errHandler(errs["space"], "Set org to lower starting address.")
		}
		var tmp []byte
		curLine = i
		if !inst.isComment {
			if inst.kind == "rel" { // handle relative addressing
				tmp = append(tmp, inst.opcode)
				var tmpAddr = [2]byte{inst.opHighByte, inst.opLowByte}
				var intTmpAddr = hexToInt(tmpAddr)
				if intTmpAddr > PC { // relative branch is positive
					diff := intTmpAddr - (PC + 2)
					if diff > 127 {
						errHandler(errs["relative"], "Positive offset greater than 127.")
					} else {
						if diff <= 255 {
							tmp = append(tmp, intToHex(diff)[0]) // low byte
						} else {
							tmp = append(tmp, intToHex(diff)[1]) // low byte
						}
						PC += 2
					}
				} else { // relative branch is negative
					testPrint(inst)
					diff := (PC + 1) - intTmpAddr
					diff = 255 - diff
					if diff < 127 {
						errHandler(errs["relative"], "Negative offset greater than -128.")
					} else {
						if diff <= 255 {
							tmp = append(tmp, intToHex(diff)[0]) // low byte
						} else {
							tmp = append(tmp, intToHex(diff)[1]) // low byte
						}
						PC += 2
					}
				}
			} else {
				tmp = append(tmp, inst.opcode)
				PC++
				if inst.length > 1 {
					tmp = append(tmp, inst.opLowByte)
					PC++
					if inst.length > 2 {
						tmp = append(tmp, inst.opHighByte)
						PC++
					}
				}
			}
		}
		obj = append(obj, tmp)
	}
	return obj
}

func getOrg(insts []instruction) {
	for _, inst := range insts {
		if inst.mnemonic == "org" {
			var addr = [2]byte{inst.opHighByte, inst.opLowByte}
			org = hexToInt(addr)
			break
		}
	}
}

func getSymbols(insts []instruction) {
	var PC int = org
	for i, inst := range insts {
		curLine = i
		if inst.label != "" && inst.kind != "pse" {
			var tmp symbol
			tmp.label = inst.label
			tmp.intAddr = PC
			tmpAddr := intToHex(tmp.intAddr)
			if tmp.intAddr <= 255 {
				tmp.addLowByte = tmpAddr[0]
			} else {
				tmp.addHighByte = tmpAddr[0]
				tmp.addLowByte = tmpAddr[1]
			}
			if symbolExists(tmp.label) {
				errHandler(errs["duplicatesym"])
			} else {
				symbols = append(symbols, tmp)
			}
		}
		if !inst.isComment && inst.kind != "pse" {
			PC += inst.length
		}
	}
}

func printAssembly(lines []string, obj [][]byte) {
	// addr | sym | ops | line | file
	// 5	  7	    10    7      no limit
	var symi int = 0
	var PC int = org
	var PCStart int = PC
	printAtWidth("\nAssembly Listing ", 75, "=")
	fmt.Println()
	for i, line := range obj {
		if len(line) > 0 {
			printAtWidth(fmt.Sprintf("%04X", PC), 5)
			// PC += len(line)
		} else {
			printAtWidth("", 5)
		}
		if len(symbols) > 0 && symbols[symi].intAddr == PC && len(line) > 0 {
			printAtWidth(symbols[symi].label, 7)
			if len(symbols)-1 > symi {
				symi++
			}
		} else {
			printAtWidth("", 7)
		}
		var tmp string
		for _, op := range line {
			tmp += fmt.Sprintf("%02X ", op)
		}
		printAtWidth(tmp, 10)
		fmt.Print("| ")
		printAtWidth(strconv.Itoa(i+1), 7)
		fmt.Println(lines[i])
		if len(line) > 0 {
			PC += len(line)
		}
	}
	fmt.Printf("\nObject will fill from $%04X through $%04X. ($%04X bytes)\n", PCStart, PC-1, PC-PCStart)
}

func printSymbolTable() {
	var colWidth int = 60
	var runWidth int = 0
	printAtWidth("\nSymbol Table ", 75, "=")
	fmt.Println()
	var tmp []string
	for _, symbol := range symbols {
		tmp = append(tmp, symbol.label)
	}
	sort.Strings(tmp)
	for _, ssymbol := range tmp {
		for _, symbol := range symbols {
			if symbol.label == ssymbol {
				if runWidth > colWidth {
					fmt.Print("\n")
					runWidth = 0
				}
				runWidth += printAtWidth(symbol.label, 8)
				runWidth += printAtWidth(fmt.Sprintf("$%04X", symbol.intAddr), 12)
			}
		}
	}
}

func saveFile(filename string, obj [][]byte) {
	var tmp []byte
	for _, line := range obj {
		for _, op := range line {
			tmp = append(tmp, op)
		}
	}
	e := ioutil.WriteFile(filename, tmp, 0644)
	if e != nil {
		errHandler(errs["file"])
	}
	fmt.Println("\n\nWrote " + strconv.Itoa(len(tmp)) + " bytes to " + filename + ".")
}

func isMnemonic(str string) bool {
	str = strings.ToLower(str)
	_, ok := mnemonics[str]
	if ok {
		return true
	}
	_, ok = pseudoOps[str]
	if ok {
		return true
	}
	return false
}

func hexToInt(addr [2]byte) int {
	tmp := fmt.Sprintf("%02x%02x", addr[0], addr[1])
	o, e := strconv.ParseInt(tmp, 16, 64)
	if e != nil {
		errHandler(errs["conversion"], "Error converting hex to int.")
	}
	return int(o)
}

func intToHex(addr int) []byte {
	var str string
	if addr <= 255 {
		str = fmt.Sprintf("%02x", addr)
	} else {
		str = fmt.Sprintf("%04x", addr)
	}
	tmp, _ := hex.DecodeString(str)
	return tmp
}

func symbolExists(sym string) bool {
	for _, symbol := range symbols {
		if sym == symbol.label {
			return true
		}
	}
	return false
}

func printAtWidth(str string, wid int, filler ...string) (length int) {
	var tmp string
	var fill string = " "
	if len(filler) > 0 {
		fill = filler[0]
	}
	if len(str) < wid {
		tmp = fmt.Sprint(str + strings.Repeat(fill, wid-len(str)))
	} else {
		tmp = fmt.Sprint(str[:wid])
	}
	fmt.Print(tmp)
	return len(tmp)
}

func errHandler(err []string, deets ...string) {
	color.FgRed.Print("\nERROR ")
	if len(lines) == 0 {
		color.FgDefault.Println("[preproc]")
	} else {
		color.FgDefault.Print("[line " + strconv.Itoa(curLine+1) + "] ")
		fmt.Println(strings.Split(strings.Split(strings.TrimSpace(lines[curLine]), "*")[0], ";")[0])
	}
	fmt.Println(err[1])
	if len(deets) > 0 {
		fmt.Println(deets[0])
	}
	if !continueOnError {
		os.Exit(1)
	}
}

// For error handling
var errs = map[string][]string{
	"conversion":   {"Hex to byte", "Could not complete conversion."},
	"duplicatesym": {"Duplicate symbol", "The label already exists in the symbol table."},
	"file":         {"File I/O", "Could not read or write to file."},
	"label":        {"Label", "Operand too short or cannot parse label."},
	"labelLength":  {"Label", "Label must not exceed " + strconv.Itoa(maxLabelLength) + " chars."},
	"length":       {"Address length", "Address length does not match opcode."},
	"mnemonic":     {"Mnemonic", "Could not find a valid mnemonic."},
	"nofile":       {"File I/O", "No file specified."},
	"opcode":       {"Opcode", "Invalid mnemonic/operand combination."},
	"operand":      {"Operand", "The operand is ill formed."},
	"org":          {"Org", "Pseudo-op address could not be determined."},
	"parser":       {"Parser", "Could not parse line successfully."},
	"relative":     {"Branching", "Relative address is out of range."},
	"space":        {"Memory", "Object will not fit in address space."},
	"symbol":       {"Symbol", "Could not determine symbol address."},
	"toomanyargs":  {"Arguments", "Too many arguments on command line"},
	"unknownsym":   {"Symbol", "Symbol not defined."}}

func removePathFileExtension(path string) (newpath string) {
	slash_chk := strings.Split(path, "/")
	for i, item := range slash_chk {
		if i == len(slash_chk)-1 {
			break
		}
		newpath += item + "/"
	}
	dot_chk := strings.Split(slash_chk[len(slash_chk)-1], ".")
	for i, item := range dot_chk {
		if i == len(dot_chk)-1 {
			break
		}
		newpath += item
	}
	return
}

func testPrint(inst instruction) {
	fmt.Println("Instruction")
	fmt.Println(strings.Repeat("=", 20))
	if inst.isComment && inst.kind != "pse" {
		fmt.Println("Comment line")
	} else {
		fmt.Println("Mnemonic: " + inst.mnemonic)
		fmt.Println("Kind:     " + inst.kind)
		fmt.Printf("Length    %x\n", inst.length)
		fmt.Printf("Opcode:   %x\n", inst.opcode)
		fmt.Printf("Add HB    %x\n", inst.opHighByte)
		fmt.Printf("Add LB    %x\n", inst.opLowByte)
		fmt.Println("Label:    " + inst.label)
	}
}
