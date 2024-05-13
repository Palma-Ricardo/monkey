package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack        []object.Object
	stackPointer int
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:        make([]object.Object, StackSize),
		stackPointer: 0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPointer == 0 {
		return nil
	}

	return vm.stack[vm.stackPointer-1]
}

func (vm *VM) Run() error {
	for instructionPointer := 0; instructionPointer < len(vm.instructions); instructionPointer++ {
		op := code.Opcode(vm.instructions[instructionPointer])

		switch op {
		case code.OpConstant:
			constantIndex := code.ReadUint16(vm.instructions[instructionPointer+1:])
			instructionPointer += 2

			error := vm.push(vm.constants[constantIndex])
			if error != nil {
				return error
			}
		}
	}

	return nil
}

func (vm *VM) push(obj object.Object) error {
	if vm.stackPointer >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.stackPointer] = obj
	vm.stackPointer++

	return nil
}
