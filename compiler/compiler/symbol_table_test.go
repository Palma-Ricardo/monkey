package compiler

import "testing"

func TestDefine(tester *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
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

	firstLocal := NewEnclosedSymbolTable(global)

	c := firstLocal.Define("c")
	if c != expected["c"] {
		tester.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}

	d := firstLocal.Define("d")
	if d != expected["d"] {
		tester.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != expected["e"] {
		tester.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}

	f := secondLocal.Define("f")
	if f != expected["f"] {
		tester.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(tester *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, symbol := range expected {
		result, ok := local.Resolve(symbol.Name)
		if !ok {
			tester.Errorf("name %s not resolvable", symbol.Name)
			continue
		}

		if result != symbol {
			tester.Errorf("expected %s to resolve to %+v, got=%+v", symbol.Name, symbol, result)
		}
	}
}

func TestResolveNestedLocal(tester *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, testcase := range tests {
		for _, symbol := range testcase.expectedSymbols {
			result, ok := testcase.table.Resolve(symbol.Name)
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
}

func TestDefineResolveBuiltins(tester *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for index, value := range expected {
		global.DefineBuiltin(index, value.Name)
	}

	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, symbol := range expected {
			result, ok := table.Resolve(symbol.Name)
			if !ok {
				tester.Errorf("name %s not resolvable", symbol.Name)
				continue
			}

			if result != symbol {
				tester.Errorf("expected %s to resolve to %+v, got=%+v", symbol.Name, symbol, result)
			}
		}
	}
}

func TestResolveFree(tester *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, testcase := range tests {
		for _, symbol := range testcase.expectedSymbols {
			result, ok := testcase.table.Resolve(symbol.Name)
			if !ok {
				tester.Errorf("name %s not resolvable", symbol.Name)
				continue
			}
			if result != symbol {
				tester.Errorf("expected %s to resolve to %+v, got=%+v",
					symbol.Name, symbol, result)
			}
		}
		if len(testcase.table.FreeSymbols) != len(testcase.expectedFreeSymbols) {
			tester.Errorf("wrong number of free symbols. got=%d, want=%d",
				len(testcase.table.FreeSymbols), len(testcase.expectedFreeSymbols))
			continue
		}
		for i, sym := range testcase.expectedFreeSymbols {
			result := testcase.table.FreeSymbols[i]
			if result != sym {
				tester.Errorf("wrong free symbol. got=%+v, want=%+v",
					result, sym)
			}
		}
	}
}

func TestResolveUnresolvableFree(tester *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, symbol := range expected {
		result, ok := secondLocal.Resolve(symbol.Name)
		if !ok {
			tester.Errorf("name %s not resolvable", symbol.Name)
			continue
		}
		if result != symbol {
			tester.Errorf("expected %s to resolve to %+v, got=%+v",
				symbol.Name, symbol, result)
		}
	}

	expectedUnresolvable := []string{
		"b",
		"d",
	}

	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)
		if ok {
			tester.Errorf("name %s resolved, but was expected not to", name)
		}
	}
}
