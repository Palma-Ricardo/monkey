package compiler

import "testing"

func TestDefine(tester *testing.T) {
	expected := map[string]Symbol{
		"a": Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		"b": Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")
	if a != expected["a"] {
		tester.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	b := global.Define("b")
	if b != expected["b"] {
		tester.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}
}

func TestResolveGlobal(tester *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, symbol := range expected {
		result, ok := global.Resolve(symbol.Name)
		if !ok {
			tester.Errorf("name %s not resolvable", symbol.Name)
			continue
		}
		if result != symbol {
			tester.Errorf("expected %s to resolve to %+v, got=%+v",
				symbol.Name, symbol, result)
		}
	}
}
