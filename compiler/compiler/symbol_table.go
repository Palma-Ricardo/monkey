package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
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

	FreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	store := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: store, FreeSymbols: free}
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

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	symbol := Symbol{Name: original.Name, Index: len(st.FreeSymbols) - 1}
	symbol.Scope = FreeScope

	st.store[original.Name] = symbol
	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	object, ok := st.store[name]
	if !ok && st.Outer != nil {
		object, ok = st.Outer.Resolve(name)
		if !ok {
			return object, ok
		}

		if object.Scope == GlobalScope || object.Scope == BuiltinScope {
			return object, ok
		}

		free := st.defineFree(object)
		return free, true
	}

	return object, ok
}

func (st *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	st.store[name] = symbol

	return symbol
}
