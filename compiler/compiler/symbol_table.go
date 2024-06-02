package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store               map[string]Symbol
	numberOfDefinitions int
}

func NewSymbolTable() *SymbolTable {
	store := make(map[string]Symbol)
	return &SymbolTable{store: store}
}

func (st *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: st.numberOfDefinitions, Scope: GlobalScope}
	st.store[name] = symbol
	st.numberOfDefinitions++

	return symbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	object, ok := st.store[name]
	return object, ok
}
