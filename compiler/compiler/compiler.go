package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
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

		afterConsequencePos := len(c.instructions)
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

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
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
	positionOfNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, instruction...)
	return positionOfNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, position int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: position}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	return c.lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[position+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPosition int, operand int) {
	op := code.Opcode(c.instructions[opPosition])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPosition, newInstruction)
}
