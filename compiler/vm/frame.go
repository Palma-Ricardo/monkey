package vm

import (
	"monkey/code"
	"monkey/object"
)

type Frame struct {
	fn                 *object.CompiledFunction
	instructionPointer int
	basePointer        int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	frame := &Frame{
		fn:                 fn,
		instructionPointer: -1,
		basePointer:        basePointer,
	}

	return frame
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
