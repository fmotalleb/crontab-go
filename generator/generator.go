// Package generator holds logic behind the generic generators
package generator

import "go.uber.org/zap"

type Func[I any, O any] = func(*zap.Logger, I) (O, bool)

type Core[I any, O any] struct {
	generators []Func[I, O]
}

func New[I any, O any]() *Core[I, O] {
	return &Core[I, O]{generators: []Func[I, O]{}}
}

func (g *Core[I, O]) Register(generator Func[I, O]) {
	g.generators = append(g.generators, generator)
}

func (g *Core[I, O]) Get(log *zap.Logger, input I) (O, bool) {
	for _, generator := range g.generators {
		item, ok := generator(log, input)
		if ok {
			return item, true
		}
	}
	var empty O
	return empty, false
}
