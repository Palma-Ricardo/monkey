package compiler

import (
	"monkey/ast"
	"monkey/code"
	"monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
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

	case *ast.InfixExpression:
		error := c.Compile(node.Left)
		if error != nil {
			return error
		}

		error = c.Compile(node.Right)
		if error != nil {
			return error
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
        c.emit(code.OpConstant, c.addConstant(integer))
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
	return position
}

func (c *Compiler) addInstruction(instruction []byte) int {
	positionOfNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, instruction...)
	return positionOfNewInstruction
}
