package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"monkey/ast"
	"monkey/code"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJECT        = "INTEGER"
	BOOLEAN_OBJECT        = "BOOLEAN"
	NULL_OBJECT           = "NULL"
	RETURN_VALUE_OBJECT   = "RETURN_VALUE"
	ERROR_OBJECT          = "ERROR"
	FUNCTION_OBJECT       = "FUNCTION"
	STRING_OBJECT         = "STRING"
	BUILTIN_OBJECT        = "BUILTIN"
	ARRAY_OBJECT          = "ARRAY"
	HASH_OBJECT           = "HASH"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
	CLOSURE_OBJ           = "CLOSURE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
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

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJECT }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJECT }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, element := range a.Elements {
		elements = append(elements, element.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	hasher := fnv.New64a()
	hasher.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: hasher.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJECT }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type Closure struct {
	Fn   *CompiledFunction
	Free []Object
}

func (cl *Closure) Type() ObjectType { return CLOSURE_OBJ }
func (cl *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", cl)
}
