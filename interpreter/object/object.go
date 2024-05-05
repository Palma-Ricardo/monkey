package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJECT      = "INTEGER"
	BOOLEAN_OBJECT      = "BOOLEAN"
	NULL_OBJECT         = "NULL"
	RETURN_VALUE_OBJECT = "RETURN_VALUE"
	ERROR_OBJECT        = "ERROR"
	FUNCTION_OBJECT     = "FUNCTION"
	STRING_OBJECT       = "STRING"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (integer *Integer) Type() ObjectType { return INTEGER_OBJECT }
func (integer *Integer) Inspect() string  { return fmt.Sprintf("%d", integer.Value) }

type Boolean struct {
	Value bool
}

func (boolean *Boolean) Type() ObjectType { return BOOLEAN_OBJECT }
func (boolean *Boolean) Inspect() string  { return fmt.Sprintf("%t", boolean.Value) }

type Null struct{}

func (null *Null) Type() ObjectType { return NULL_OBJECT }
func (null *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJECT }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (err *Error) Type() ObjectType { return ERROR_OBJECT }
func (err *Error) Inspect() string  { return "ERROR: " + err.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() ObjectType { return FUNCTION_OBJECT }
func (fn *Function) Inspect() string {
	var out bytes.Buffer

	parameters := []string{}
	for _, parameter := range fn.Parameters {
		parameters = append(parameters, parameter.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(parameters, ", "))
	out.WriteString(") {\n")
	out.WriteString(fn.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (str *String) Type() ObjectType { return STRING_OBJECT }
func (str *String) Inspect() string  { return str.Value }
