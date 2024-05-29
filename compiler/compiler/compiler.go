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
		c.emit(code.OpPop)

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
	return position
}

func (c *Compiler) addInstruction(instruction []byte) int {
	positionOfNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, instruction...)
	return positionOfNewInstruction
}
