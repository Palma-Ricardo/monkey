package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"sort"
)

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	for index, value := range object.Builtins {
		symbolTable.DefineBuiltin(index, value.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

func NewWithState(st *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = st
	compiler.constants = constants

	return compiler
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			error := c.Compile(statement)
			if error != nil {
				return error
			}
		}

	case *ast.ExpressionStatement:
		error := c.Compile(node.Expression)
		if error != nil {
			return error
		}
		c.emit(code.OpPop)

	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			error := c.Compile(statement)
			if error != nil {
				return error
			}
		}

	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)
		error := c.Compile(node.Value)
		if error != nil {
			return error
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}

	case *ast.ReturnStatement:
		error := c.Compile(node.ReturnValue)
		if error != nil {
			return error
		}

		c.emit(code.OpReturnValue)

	case *ast.InfixExpression:
		if node.Operator == "<" {
			error := c.Compile(node.Right)
			if error != nil {
				return error
			}

			error = c.Compile(node.Left)
			if error != nil {
				return error
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		error := c.Compile(node.Left)
		if error != nil {
			return error
		}

		error = c.Compile(node.Right)
		if error != nil {
			return error
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">":
			c.emit(code.OpGreaterThan)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PrefixExpression:
		error := c.Compile(node.Right)
		if error != nil {
			return error
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		error := c.Compile(node.Condition)
		if error != nil {
			return error
		}

		jumpNotTruePos := c.emit(code.OpJumpNotTrue, 9999)

		error = c.Compile(node.Consequence)
		if error != nil {
			return error
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruePos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			error := c.Compile(node.Alternative)
			if error != nil {
				return error
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.IndexExpression:
		error := c.Compile(node.Left)
		if error != nil {
			return error
		}

		error = c.Compile(node.Index)
		if error != nil {
			return error
		}

		c.emit(code.OpIndex)

	case *ast.CallExpression:
		error := c.Compile(node.Function)
		if error != nil {
			return error
		}

		for _, argument := range node.Arguments {
			error := c.Compile(argument)
			if error != nil {
				return error
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			error := c.Compile(element)
			if error != nil {
				return error
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for key := range node.Pairs {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			error := c.Compile(key)
			if error != nil {
				return error
			}
			error = c.Compile(node.Pairs[key])
			if error != nil {
				return error
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.FunctionLiteral:
		c.enterScope()

		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		for _, parameter := range node.Parameters {
			c.symbolTable.Define(parameter.Value)
		}

		error := c.Compile(node.Body)
		if error != nil {
			return error
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numberOfDefinitions
		instructions := c.leaveScope()

		for _, symbol := range freeSymbols {
			c.loadSymbol(symbol)
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.loadSymbol(symbol)
	}

	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	instruction := code.Make(op, operands...)
	position := c.addInstruction(instruction)
	c.setLastInstruction(op, position)
	return position
}

func (c *Compiler) addInstruction(instruction []byte) int {
	positionOfNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), instruction...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return positionOfNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, position int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: position}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	instructions := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		instructions[position+i] = newInstruction[i]
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) changeOperand(opPosition int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPosition])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPosition, newInstruction)
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	return instructions
}

func (c *Compiler) loadSymbol(sym Symbol) {
	switch sym.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, sym.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, sym.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, sym.Index)
	case FreeScope:
		c.emit(code.OpGetFree, sym.Index)
    case FunctionScope:
        c.emit(code.OpCurrentClosure)
	}
}
