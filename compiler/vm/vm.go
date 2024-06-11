package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65535

type VM struct {
	constants    []object.Object
	instructions code.Instructions
	globals      []object.Object

	stack        []object.Object
	stackPointer int
}

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		globals:      make([]object.Object, GlobalsSize),

		stack:        make([]object.Object, StackSize),
		stackPointer: 0,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, store []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = store

	return vm
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

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[instructionPointer+1:])
			instructionPointer += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[instructionPointer+1:])
			instructionPointer += 2

			error := vm.push(vm.globals[globalIndex])
			if error != nil {
				return error
			}

		case code.OpArray:
			numberElements := int(code.ReadUint16(vm.instructions[instructionPointer+1:]))
			instructionPointer += 2

			array := vm.buildArray(vm.stackPointer-numberElements, vm.stackPointer)
			vm.stackPointer = vm.stackPointer - numberElements

			error := vm.push(array)
			if error != nil {
				return error
			}

		case code.OpHash:
			numberElements := int(code.ReadUint16(vm.instructions[instructionPointer+1:]))
			instructionPointer += 2

			hash, error := vm.buildHash(vm.stackPointer-numberElements, vm.stackPointer)
			if error != nil {
				return error
			}

			vm.stackPointer = vm.stackPointer - numberElements

			error = vm.push(hash)
			if error != nil {
				return error
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			error := vm.executeBinaryOperation(op)
			if error != nil {
				return error
			}

		case code.OpTrue:
			error := vm.push(True)
			if error != nil {
				return error
			}

		case code.OpFalse:
			error := vm.push(False)
			if error != nil {
				return error
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			error := vm.executeComparison(op)
			if error != nil {
				return error
			}

		case code.OpBang:
			error := vm.executeBangOperator()
			if error != nil {
				return error
			}

		case code.OpMinus:
			error := vm.executeMinusOperator()
			if error != nil {
				return error
			}

		case code.OpJump:
			position := int(code.ReadUint16(vm.instructions[instructionPointer+1:]))
			instructionPointer = position - 1

		case code.OpJumpNotTrue:
			position := int(code.ReadUint16(vm.instructions[instructionPointer+1:]))
			instructionPointer += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				instructionPointer = position - 1
			}

		case code.OpNull:
			error := vm.push(Null)
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

	switch {
	case leftType == object.INTEGER_OBJECT && rightType == object.INTEGER_OBJECT:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING_OBJECT && rightType == object.STRING_OBJECT:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
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

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJECT && right.Type() == object.INTEGER_OBJECT {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False, Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJECT {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}

	return False
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for index := startIndex; index < endIndex; index += 2 {
		key := vm.stack[index]
		value := vm.stack[index+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}
