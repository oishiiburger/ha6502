// Hobbyist's Assembler for 6502 microprocessors

package main

import (
	"encoding/hex"
	"fmt"
	"os"
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
	mnemonic    string
	kind        string
	opcode      byte
	label       string
	addLowByte  byte
	addHighByte byte
	opLowByte   byte
	opHighByte  byte
	isComment   bool
}

const comchars string = ";*"
const modchars string = "#$"

var curLine int = 55 // set to 0 when not testing

var errs = map[string][]string{
	"conversion": {"Hex to byte", "Could not complete conversion."},
	"mnemonic":   {"Mnemonic", "Could not find a valid mnemonic."},
	"opcode":     {"Opcode", "Invalid mnemonic/operand combination."},
	"operand":    {"Operand", "The operand is ill formed."},
	"parser":     {"Parser", "Could not parse line successfully."}}

func main() {
	fmt.Println(info["title"])
	testPrint(assignOpcode(parseLine("")))
}

func parseLine(line string) (cur instruction) {
	if strings.ContainsAny(line, comchars) { // strip comments
		for _, char := range comchars {
			line = strings.Split(line, string(char))[0]
		}
		if len(line) == 0 { // if the entire line is a comment
			cur.isComment = true
			return cur
		}
	}
	cur.isComment = false
	lineArr := strings.Fields(line)
	if len(lineArr) > 0 && len(lineArr) < 4 {
		if onPeriphery(lineArr[0], ":", true) {
			cur.label = strings.ReplaceAll(lineArr[0], ":", "")
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

	return cur
}

func parseOperand(op string, inst instruction) instruction {
	op = strings.ToLower(op)
	if len(op) < 2 {
		errHandler(errs["operand"], "Operand is too short.")
	}
	// if strings.ContainsAny(op, symbols) { // label check
	// 	// do label things
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
			inst.kind = "zp"
		} else if len(tmp) == 4 {
			inst.kind = "abs"
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
		if len(label) == 0 {
			inst.opHighByte = bytes[0]
			inst.opLowByte = bytes[1]
		} else {
			inst.addHighByte = bytes[0]
			inst.addLowByte = bytes[1]
		}
	} else if len(bytes) == 1 {
		if len(label) == 0 {
			inst.opLowByte = bytes[0]
		} else {
			inst.addLowByte = bytes[0]
		}
	} else {
		errHandler(errs["conversion"], "Wrong length of address.")
	}
	return inst
}

func assignOpcode(inst instruction) instruction {
	switch inst.kind {
	case "zop":
		inst.opcode = opZop[inst.mnemonic]
	case "imm":
		inst.opcode = opImm[inst.mnemonic]
	default:
		errHandler(errs["opcode"])
	}
	return inst
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
		fmt.Printf("Op HB     %x\n", inst.addHighByte)
		fmt.Printf("Op LB     %x\n", inst.addLowByte)
	}
}
