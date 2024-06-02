package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	index := 0
	for index < len(ins) {
		definition, error := Lookup(ins[index])
		if error != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", error)
			continue
		}

		operands, read := ReadOperands(definition, ins[index+1:])

		fmt.Fprintf(&out, "%04d %s\n", index, ins.fmtInstruction(definition, operands))

		index += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(definition *Definition, operands []int) string {
	operandCount := len(definition.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return definition.Name
	case 1:
		return fmt.Sprintf("%s %d", definition.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", definition.Name)
}

type Opcode byte

const (
	OpConstant Opcode = iota
	OpNull

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpBang
	OpMinus

	OpTrue
	OpFalse

	OpEqual
	OpNotEqual
	OpGreaterThan

	OpJumpNotTrue
	OpJump

	OpSetGlobal
	OpGetGlobal

	OpPop
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpNull:     {"OpNull", []int{}},

	OpAdd:   {"OpAdd", []int{}},
	OpSub:   {"OpSub", []int{}},
	OpMul:   {"OpMul", []int{}},
	OpDiv:   {"OpDiv", []int{}},
	OpBang:  {"OpBang", []int{}},
	OpMinus: {"OpMinus", []int{}},

	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},

	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},

	OpJumpNotTrue: {"OpJumpNotTrue", []int{2}},
	OpJump:        {"OpJump", []int{2}},

    OpSetGlobal: {"OpSetGlobal", []int{2}},
	OpGetGlobal: {"OpGetGlobal", []int{2}},

	OpPop: {"OpPop", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	definition, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return definition, nil
}

func Make(op Opcode, operands ...int) []byte {
	definition, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, width := range definition.OperandWidths {
		instructionLen += width
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for index, operand := range operands {
		width := definition.OperandWidths[index]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		}
		offset += width
	}

	return instruction
}

func ReadOperands(definition *Definition, instruction Instructions) ([]int, int) {
	operands := make([]int, len(definition.OperandWidths))
	offset := 0

	for index, width := range definition.OperandWidths {
		switch width {
		case 2:
			operands[index] = int(ReadUint16(instruction[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(instruction Instructions) uint16 {
	return binary.BigEndian.Uint16(instruction)
}
