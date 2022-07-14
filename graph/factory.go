package graph

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/module"
)

func GraphFromComponentMatcher(matcher module.ComponentMatcher) (Graph, error) {
	gb := NewGraphBuilder()
	matcher.All().Each(func(c module.Component) {
		vextexOfComponent(c, matcher, gb)
	})

	err := gb.Validate()
	if err != nil {
		return nil, err
	}

	return gb, nil
}

func vextexOfComponent(c module.Component, matcher module.ComponentMatcher, gb GraphBuilder) Vertex {
	cVertex := gb.VertexBy(c)
	if cVertex != nil {
		return cVertex
	}

	funcVertex := gb.AddFuncVertex(c.Provider().FuncVal(), c.Provider())

	// here must process component vertices before parameter edges.
	// if not it will end in "infinity loop"
	c.Provider().Components().Each(func(com module.Component) {
		comVertex := gb.AddIndexedVertex(c.Index(), com)
		gb.AddEdge(comVertex, funcVertex, -1)
	})

	c.Provider().Parameters().Each(func(p module.Parameter) {
		pVertex := vextexOfParameter(p, matcher, gb)
		gb.AddEdge(funcVertex, pVertex, p.Index())
	})

	return gb.VertexBy(c)
}

func vextexOfParameter(p module.Parameter, matcher module.ComponentMatcher, gb GraphBuilder) Vertex {
	var cl = matcher.Match(p, p.Provider().Scope())
	var pVertex Vertex

	// TODO: filter out provider himself?
	if p.IsCollector() {
		pVertex = gb.AddArrayVertex(nil)
		cl.Each(func(com module.Component) {
			itemVertex := vextexOfComponent(com, matcher, gb)
			gb.AddEdge(pVertex, itemVertex, -1)
		})
	} else {
		coms := module.ArrayOfComponents(cl)
		if len(coms) == 0 {
			if p.Optional() {
				val := reflect.Zero(p.Type())
				pVertex = gb.AddValueVertex(val, nil)
			} else {
				// TODO: add info
				pVertex = gb.AddErrorVertex(fmt.Errorf("missing dependency"), nil)
			}
		} else if len(coms) == 1 {
			pVertex = vextexOfComponent(coms[0], matcher, gb)
		} else {
			pVertex := gb.AddOneOfVertex(nil)
			cl.Each(func(com module.Component) {
				candicateVertex := vextexOfComponent(com, matcher, gb)
				gb.AddEdge(pVertex, candicateVertex, -1)
			})
		}
	}

	return pVertex
}
