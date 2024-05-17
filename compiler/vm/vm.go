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

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.stackPointer]
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
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			error := vm.executeBinaryOperation(op)
			if error != nil {
				return error
			}

		case code.OpPop:
			vm.pop()
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

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.stackPointer-1]
	vm.stackPointer--
	return obj
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJECT && rightType == object.INTEGER_OBJECT {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}
