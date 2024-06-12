package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store               map[string]Symbol
	numberOfDefinitions int
}

func NewSymbolTable() *SymbolTable {
	store := make(map[string]Symbol)
	return &SymbolTable{store: store}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (st *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: st.numberOfDefinitions}
	if st.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	st.store[name] = symbol
	st.numberOfDefinitions++

	return symbol
}

func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	st.store[name] = symbol

	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	object, ok := st.store[name]
	if !ok && st.Outer != nil {
		object, ok = st.Outer.Resolve(name)
		return object, ok
	}

	return object, ok
}
