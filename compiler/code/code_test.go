package code

import "testing"

func TestMake(tester *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
	}

	for _, testcase := range tests {
		instruction := Make(testcase.op, testcase.operands...)

		if len(instruction) != len(testcase.expected) {
			tester.Errorf("instruction has wrong length. want=%d, got=%d",
				len(testcase.expected), len(instruction))
		}

		for index, b := range testcase.expected {
			if instruction[index] != testcase.expected[index] {
				tester.Errorf("wrong byte at pos %d. want=%d, got=%d",
					index, b, instruction[index])
			}
		}
	}
}

func TestInstructionsString(tester *testing.T) {
	instructions := []Instructions{
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpConstant 1
0003 OpConstant 2
0006 OpConstant 65535
`

	concatenated := Instructions{}
	for _, instruction := range instructions {
		concatenated = append(concatenated, instruction...)
	}

	if concatenated.String() != expected {
		tester.Errorf("instruction wrongly formatted.\nwant=%q\ngot=%q",
			expected, concatenated.String())
	}
}

func TestReadOperands(tester *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
	}

	for _, testcase := range tests {
		instruction := Make(testcase.op, testcase.operands...)

		definition, error := Lookup(byte(testcase.op))
		if error != nil {
			tester.Fatalf("definition not found: %q\n", error)
		}

		operandsRead, numberRead := ReadOperands(definition, instruction[1:])
		if numberRead != testcase.bytesRead {
			tester.Fatalf("number wrong. want=%d, got=%d", testcase.bytesRead, numberRead)
		}

		for index, wanted := range testcase.operands {
			if operandsRead[index] != wanted {
				tester.Errorf("operand wrong. want=%d, got=%d", wanted, operandsRead[i])
			}
		}
	}
}
