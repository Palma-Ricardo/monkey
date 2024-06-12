package vm

import (
	"monkey/code"
	"monkey/object"
)

type Frame struct {
	cl                 *object.Closure
	instructionPointer int
	basePointer        int
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	frame := &Frame{
		cl:                 cl,
		instructionPointer: -1,
		basePointer:        basePointer,
	}

	return frame
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
