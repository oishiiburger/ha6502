// Hobbyist's Assembler for 6502 microprocessors

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
	"title":      "Hobbyist Assembler for 6502 microprocessors",
	"shortTitle": "ha6502",
}

var continueOnError bool = false

// Each line will get parsed into an instance of line
type instruction struct {
	mnemonic   string
	kind       string
	opcode     byte
	length     int
	label      string
	opLowByte  byte
	opHighByte byte
	isComment  bool
}

type symbol struct {
	label       string
	addLowByte  byte
	addHighByte byte
	intAddr     int
}

const comchars string = ";*"
const modchars string = "#$"
const maxLabelLength int = 6

var filename string
var ofilename string

var curLine int = 0 // set to 0 when not testing
var curInsts []instruction
var curSymbols []symbol
var curObj [][]byte
var curOrg int = 0

var errs = map[string][]string{
	"conversion":  {"Hex to byte", "Could not complete conversion."},
	"file":        {"File I/O", "Could not read or write to file."},
	"labelLength": {"Label", "Label must not exceed " + strconv.Itoa(maxLabelLength) + " chars."},
	"mnemonic":    {"Mnemonic", "Could not find a valid mnemonic."},
	"opcode":      {"Opcode", "Invalid mnemonic/operand combination."},
	"operand":     {"Operand", "The operand is ill formed."},
	"org":         {"Org", "Pseudo-op address could not be determined."},
	"parser":      {"Parser", "Could not parse line successfully."},
	"symbol":      {"Symbol", "Could not determine symbol address."}}

func main() {
	fmt.Println(info["title"])
	// testPrint(assignOpcode(parseLine("test:     jmp (45)")))
	filename = "./files/test.s"
	ofilename = "./files/out.o"
	lines := loadFile(filename)
	for i, line := range lines {
		curLine = i
		curInsts = append(curInsts, parseLine(line))
	}
	curObj = asmObject(curInsts)
	curOrg = setOrg(curInsts)
	curSymbols = getSymbols(curInsts, curObj, curOrg)
	printAssembly(lines, curObj, curSymbols, curOrg)
	printSymbolTable(curSymbols)
	saveFile(ofilename, curObj)
	// passOne()
	// passTwo()
}

func loadFile(filename string) (lines []string) {
	file, e := ioutil.ReadFile(filename)
	if e != nil {
		errHandler(errs["file"])
	}
	lines = strings.Split(string(file), "\n")
	return lines
}

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
		if onPeriphery(lineArr[0], ":", true) {
			cur.label = strings.ReplaceAll(lineArr[0], ":", "")
			if len(cur.label) > maxLabelLength {
				errHandler(errs["labelLength"])
			}
			if len(lineArr) > 1 {
				if isMnemonic(lineArr[1]) { // mnemonic, label
					cur.mnemonic = lineArr[1]
					if len(lineArr) == 2 {
						cur.kind = "zop"
					}
				} else {
					errHandler(errs["mnemonic"])
				}
				if len(lineArr) > 2 {
					cur = parseOperand(lineArr[2], cur)
				}
			}
		} else {
			if isMnemonic(lineArr[0]) { // mnemonic, no label
				cur.mnemonic = lineArr[0]
				if len(lineArr) == 1 {
					cur.kind = "zop"
				}
			} else {
				errHandler(errs["mnemonic"])
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
	if len(op) < 2 {
		errHandler(errs["operand"], "Operand is too short.")
	}
	// if strings.ContainsAny(op, symbols) { // label check
	// 	// do label things in pass 2
	// }
	var tmp string
	if onPeriphery(op, "(", false) { // indirect indexed and indexed indirect
		if strings.ContainsAny(string(op[1]), "#") {
			errHandler(errs["operand"], "Operand cannot be immediate.")
		}
		if len(op) < 4 {
			errHandler(errs["operand"], "Operand is too short.")
		}
		if onPeriphery(op, ",x)", true) {
			inst.kind = "zpxi"
			op = removePeriphery(op, "(", ",x)")
		} else if len(op) > 2 && onPeriphery(op, "),y", true) {
			inst.kind = "zpiy"
			op = removePeriphery(op, "(", "),y")
		} else if onPeriphery(op, ")", true) {
			inst.kind = "ind"
			op = removePeriphery(op, "(", ")")
			tmp = removeChars(op, modchars)
			if len(tmp) != 4 && len(tmp) != 2 {
				errHandler(errs["operand"], "Address not of expected length.")
			}
		} else {
			errHandler(errs["operand"])
		}
		inst = parseAddress(op, inst)
	} else if len(op) > 2 && onPeriphery(op, ",x", true) {
		if onPeriphery(op, "#", false) {
			errHandler(errs["operand"], "Operand cannot be immediate.")
		}
		tmp = removeChars(removePeriphery(op, "", ",x"), modchars)
		if len(tmp) == 2 {
			inst.kind = "zpx"
		} else if len(tmp) == 4 {
			inst.kind = "absx"
		} else {
			errHandler(errs["operand"], "Bad address.")
		}
		inst = parseAddress(tmp, inst)
	} else if len(op) > 2 && onPeriphery(op, ",y", true) {
		if onPeriphery(op, "#", false) {
			errHandler(errs["operand"], "Operand cannot be immediate.")
		}
		tmp = removeChars(removePeriphery(op, "", ",y"), modchars)
		if len(tmp) != 4 {
			errHandler(errs["operand"], "Bad address.")
		} else {
			inst.kind = "absy"
		}
		inst = parseAddress(tmp, inst)
	} else if onPeriphery(op, "#", false) {
		tmp = removeChars(op, modchars)
		if len(tmp) != 2 {
			errHandler(errs["operand"], "Immediate value must be one byte.")
		} else {
			inst.kind = "imm"
		}
		inst = parseAddress(tmp, inst)
	} else {
		tmp = removeChars(op, modchars)
		if len(tmp) == 2 {
			_, ok := opRel[inst.mnemonic]
			if ok {
				inst.kind = "rel"
			} else {
				inst.kind = "zp"
			}
		} else if len(tmp) == 4 {
			_, ok := opRel[inst.mnemonic]
			if ok {
				inst.kind = "rel"
			} else {
				inst.kind = "abs"
			}
		} else {
			errHandler(errs["operand"], "Bad address.")
		}
		inst = parseAddress(tmp, inst)
	}
	return inst
}

func parseAddress(addr string, inst instruction, label ...bool) instruction {
	for _, char := range modchars {
		addr = strings.ReplaceAll(addr, string(char), "")
	}
	if !(len(addr) == 2 || len(addr) == 4) {
		errHandler(errs["operand"], "Address is ill formed.")
	}
	bytes, e := hex.DecodeString(addr)
	if e != nil {
		errHandler(errs["conversion"])
	}
	if len(bytes) == 2 {
		inst.opHighByte = bytes[0]
		inst.opLowByte = bytes[1]
	} else if len(bytes) == 1 {
		inst.opLowByte = bytes[0]
	} else {
		errHandler(errs["conversion"], "Wrong length of address.")
	}
	return inst
}

func assignOpcode(inst instruction) instruction {
	if !inst.isComment {
		_, ok := pseudoOps[inst.mnemonic]
		if ok {
			inst.kind = "pse"
			inst.isComment = true // set pseudo-ops to comments
			inst.length = 3
		} else {
			switch inst.kind {
			case "zop":
				_, ok = opZop[inst.mnemonic]
				if ok {
					inst.opcode = opZop[inst.mnemonic]
					inst.length = 1
				}
			case "imm":
				_, ok = opImm[inst.mnemonic]
				if ok {
					inst.opcode = opImm[inst.mnemonic]
					inst.length = 2
				}
			case "zp":
				_, ok = opZp[inst.mnemonic]
				if ok {
					inst.opcode = opZp[inst.mnemonic]
					inst.length = 2
				}
			case "zpx":
				_, ok = opZpx[inst.mnemonic]
				if ok {
					inst.opcode = opZpx[inst.mnemonic]
					inst.length = 2
				}
			case "abs":
				_, ok = opAbs[inst.mnemonic]
				if ok {
					inst.opcode = opAbs[inst.mnemonic]
					inst.length = 3
				}
			case "absx":
				_, ok = opAbsx[inst.mnemonic]
				if ok {
					inst.opcode = opAbsx[inst.mnemonic]
					inst.length = 3
				}
			case "absy":
				_, ok = opAbsy[inst.mnemonic]
				if ok {
					inst.opcode = opAbsy[inst.mnemonic]
					inst.length = 3
				}
			case "zpxi":
				_, ok = opZpxi[inst.mnemonic]
				if ok {
					inst.opcode = opZpxi[inst.mnemonic]
					inst.length = 2
				}
			case "zpiy":
				_, ok = opZpiy[inst.mnemonic]
				if ok {
					inst.opcode = opZpiy[inst.mnemonic]
					inst.length = 2
				}
			case "ind":
				_, ok = opInd[inst.mnemonic]
				if ok {
					inst.opcode = opInd[inst.mnemonic]
					inst.length = 3
				}
			case "rel":
				_, ok = opRel[inst.mnemonic]
				if ok {
					inst.opcode = opRel[inst.mnemonic]
					inst.length = 3
				}
			default:
				errHandler(errs["opcode"])
			}
		}
		// if inst.opcode == 0 && inst.mnemonic != "brk" {
		// 	errHandler(errs["opcode"])
		// }
	}
	return inst
}

func asmObject(insts []instruction) (obj [][]byte) {
	for i, inst := range insts {
		var tmp []byte
		curLine = i
		if !inst.isComment {
			tmp = append(tmp, inst.opcode)
			if inst.length > 1 {
				tmp = append(tmp, inst.opLowByte)
				if inst.length > 2 {
					tmp = append(tmp, inst.opHighByte)
				}
			}
		}
		obj = append(obj, tmp)
	}
	return obj
}

func setOrg(insts []instruction) (org int) {
	for _, inst := range insts {
		if inst.mnemonic == "org" {
			var addr = [2]byte{inst.opHighByte, inst.opLowByte}
			return hexToInt(addr)
		}
	}
	return 0
}

func getSymbols(insts []instruction, obj [][]byte, PC int) (symbols []symbol) {
	for i, inst := range insts {
		curLine = i
		if inst.label != "" {
			var tmp symbol
			tmp.label = inst.label
			tmpAddr := [2]byte{inst.opHighByte, inst.opLowByte}
			tmp.intAddr = hexToInt(tmpAddr) + PC
			tmp.addHighByte = intToHex(tmp.intAddr)[0]
			tmp.addLowByte = intToHex(tmp.intAddr)[1]
			symbols = append(symbols, tmp)
		}
		if !inst.isComment && !(inst.kind == "pse") {
			PC += inst.length
		}
	}
	return symbols
}

func printAssembly(lines []string, obj [][]byte, symbols []symbol, PC int) {
	// addr | sym | ops | line | file
	// 5	  7	    10    7      no limit
	var symi int = 0
	printAtWidth("\nAssembly Listing ", 75, "=")
	fmt.Println()
	for i, line := range obj {
		if len(line) > 0 {
			printAtWidth(fmt.Sprintf("%04X", PC), 5)
			// PC += len(line)
		} else {
			printAtWidth("", 5)
		}
		if symbols[symi].intAddr == PC && len(line) > 0 {
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
}

func printSymbolTable(symbols []symbol) {
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
				runWidth += printAtWidth(symbol.label, 8)
				runWidth += printAtWidth(fmt.Sprintf("%04X", symbol.intAddr), 12)
				if runWidth > colWidth {
					fmt.Print("\n")
					runWidth = 0
				}
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
	fmt.Println("\nWrote " + strconv.Itoa(len(tmp)) + " bytes to " + filename + ".")
}

func onPeriphery(str string, seq string, right bool) bool {
	if len(seq) <= len(str) {
		if !right {
			if str[:len(seq)] == seq {
				return true
			}
		} else {
			if str[len(str)-len(seq):] == seq {
				return true
			}
		}
	} else {
		errHandler(errs["parser"])
	}
	return false
}

func removePeriphery(in string, l string, r string) (out string) {
	if len(l)+len(r) > len(in) {
		errHandler(errs["parser"])
	}
	return in[len(l) : len(in)-len(r)]
}

func removeChars(in string, chars string) string {
	for _, char := range chars {
		in = strings.ReplaceAll(in, string(char), "")
	}
	return in
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

func isDigit(str string) bool {
	var digits string = "0123456789"
	if strings.ContainsAny(str, digits) {
		return true
	}
	return false
}

func hexToInt(addr [2]byte) int {
	tmp := fmt.Sprintf("%02x%02x", addr[0], addr[1])
	o, e := strconv.ParseInt(tmp, 16, 16)
	if e != nil {
		errHandler(errs["org"])
	}
	return int(o)
}

func intToHex(addr int) []byte {
	str := fmt.Sprintf("%04x", addr)
	tmp, _ := hex.DecodeString(str)
	return tmp
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
	fmt.Println("\n" + strings.Repeat("=", 40))
	tmp := "ERROR: " + err[0]
	var num string
	if curLine != 0 {
		num = "Line " + strconv.Itoa(curLine)
	} else {
		num = ""
	}
	length := len(tmp) + len(num)
	if length < 40 {
		color.FgRed.Println(tmp + strings.Repeat(" ", 40-length) + num)
	} else {
		color.FgRed.Println(tmp + " " + num)
	}
	color.FgDefault.Println(err[1])
	if len(deets) > 0 {
		fmt.Println(deets[0])
	}
	fmt.Println(strings.Repeat("=", 40) + "\n")
	if !continueOnError {
		os.Exit(1)
	}
}

func testPrint(inst instruction) {
	fmt.Println("Instruction")
	fmt.Println(strings.Repeat("=", 20))
	if inst.isComment {
		fmt.Println("Comment line")
	} else {
		fmt.Println("Mnemonic: " + inst.mnemonic)
		fmt.Println("Kind:     " + inst.kind)
		fmt.Printf("Opcode:   %x\n", inst.opcode)
		fmt.Printf("Add HB    %x\n", inst.opHighByte)
		fmt.Printf("Add LB    %x\n", inst.opLowByte)
		fmt.Println("Label:    " + inst.label)
	}
}
