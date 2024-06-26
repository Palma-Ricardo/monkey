package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65535
const MaxFrames = 1024

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack        []object.Object
	stackPointer int

	frames     []*Frame
	frameIndex int
}

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,
		globals:   make([]object.Object, GlobalsSize),

		stack:        make([]object.Object, StackSize),
		stackPointer: 0,

		frames:     frames,
		frameIndex: 1,
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
	var instructionPointer int
	var instructions code.Instructions
	var op code.Opcode

	for vm.currentFrame().instructionPointer < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().instructionPointer++

		instructionPointer = vm.currentFrame().instructionPointer
		instructions = vm.currentFrame().Instructions()
		op = code.Opcode(instructions[instructionPointer])

		switch op {
		case code.OpConstant:
			constantIndex := code.ReadUint16(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 2

			error := vm.push(vm.constants[constantIndex])
			if error != nil {
				return error
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 2

			error := vm.push(vm.globals[globalIndex])
			if error != nil {
				return error
			}

		case code.OpSetLocal:
			localIndex := code.ReadUint8(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 1

			frame := vm.currentFrame()

			error := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if error != nil {
				return error
			}

		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 1

			definition := object.Builtins[builtinIndex]

			error := vm.push(definition.Builtin)
			if error != nil {
				return error
			}

		case code.OpGetFree:
			freeIndex := code.ReadUint8(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 1

			currentClosure := vm.currentFrame().cl

			error := vm.push(currentClosure.Free[freeIndex])
			if error != nil {
				return error
			}

		case code.OpArray:
			numberElements := int(code.ReadUint16(instructions[instructionPointer+1:]))
			vm.currentFrame().instructionPointer += 2

			array := vm.buildArray(vm.stackPointer-numberElements, vm.stackPointer)
			vm.stackPointer = vm.stackPointer - numberElements

			error := vm.push(array)
			if error != nil {
				return error
			}

		case code.OpHash:
			numberElements := int(code.ReadUint16(instructions[instructionPointer+1:]))
			vm.currentFrame().instructionPointer += 2

			hash, error := vm.buildHash(vm.stackPointer-numberElements, vm.stackPointer)
			if error != nil {
				return error
			}

			vm.stackPointer = vm.stackPointer - numberElements

			error = vm.push(hash)
			if error != nil {
				return error
			}

		case code.OpClosure:
			constIndex := code.ReadUint16(instructions[instructionPointer+1:])
			numFree := code.ReadUint8(instructions[instructionPointer+3:])
			vm.currentFrame().instructionPointer += 3

			error := vm.pushClosure(int(constIndex), int(numFree))
			if error != nil {
				return error
			}

		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().cl
			error := vm.push(currentClosure)
			if error != nil {
				return error
			}

		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			error := vm.executeIndexExpression(left, index)
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

		case code.OpCall:
			numArgs := code.ReadUint8(instructions[instructionPointer+1:])
			vm.currentFrame().instructionPointer += 1

			error := vm.executeCall(int(numArgs))
			if error != nil {
				return error
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.stackPointer = frame.basePointer - 1

			error := vm.push(returnValue)
			if error != nil {
				return error
			}

		case code.OpReturn:
			frame := vm.popFrame()
			vm.stackPointer = frame.basePointer - 1

			error := vm.push(Null)
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
			position := int(code.ReadUint16(instructions[instructionPointer+1:]))
			vm.currentFrame().instructionPointer = position - 1

		case code.OpJumpNotTrue:
			position := int(code.ReadUint16(instructions[instructionPointer+1:]))
			vm.currentFrame().instructionPointer += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().instructionPointer = position - 1
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

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJECT && index.Type() == object.INTEGER_OBJECT:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJECT:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.stackPointer-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.stackPointer-numArgs)
	vm.pushFrame(frame)

	vm.stackPointer = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.stackPointer-numArgs : vm.stackPointer]

	result := builtin.Fn(args...)
	vm.stackPointer = vm.stackPointer - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.stackPointer-numFree+i]
	}
	vm.stackPointer = vm.stackPointer - numFree

	closure := &object.Closure{Fn: function, Free: free}
	return vm.push(closure)
}
