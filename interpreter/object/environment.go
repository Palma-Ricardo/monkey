package object

type Environment struct {
	store map[string]Object
}

func NewEnvironment() *Environment {
	store := make(map[string]Object)
	return &Environment{store: store}
}

func (env *Environment) Get(name string) (Object, bool) {
	object, ok := env.store[name]
	return object, ok
}

func (env *Environment) Set(name string, value Object) Object {
	env.store[name] = value
	return value
}
