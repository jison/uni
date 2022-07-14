package graph

import (
	"reflect"
	"sort"

	"github.com/jison/uni/internal/errors"
)

type graph struct {
	vertexByKey     vertexMap
	verticesFromSrc map[Vertex]orderedVertexArray
	verticesToDst   map[Vertex]vertexArray
}

type vertexArray []Vertex

func (arr vertexArray) Each(f func(Vertex)) {
	for _, v := range arr {
		f(v)
	}
}

type orderedVertex struct {
	Vertex
	order uint
}

type orderedVertexArray []orderedVertex

func (arr orderedVertexArray) Each(f func(Vertex)) {
	sort.SliceStable(arr, func(i, j int) bool {
		return arr[i].order < arr[j].order
	})

	for _, v := range arr {
		f(v)
	}
}

type vertexMap map[interface{}]Vertex

func (m vertexMap) Each(f func(Vertex)) {
	for _, v := range m {
		f(v)
	}
}

var _ GraphBuilder = &graph{}

func NewGraphBuilder() GraphBuilder {
	return &graph{
		vertexByKey:     make(vertexMap, 0),
		verticesFromSrc: make(map[Vertex]orderedVertexArray, 0),
		verticesToDst:   make(map[Vertex]vertexArray, 0),
	}
}

func (g *graph) Verteies() VertexList {
	return g.vertexByKey
}

func (g *graph) VertexBy(key interface{}) Vertex {
	return g.vertexByKey[key]
}

func (g *graph) VerticesToDst(v Vertex) VertexList {
	return g.verticesToDst[v]
}

func (g *graph) VerticesFromSrc(v Vertex) VertexList {
	return g.verticesFromSrc[v]
}

func (g *graph) Provide(v Vertex, s ScopeStorage) VertexValue {
	if value, ok := s.Load(v); ok {
		return value
	}

	dependedValues := make([]VertexValue, 0)
	g.VerticesFromSrc(v).Each(func(dv Vertex) {
		dependedValue := g.Provide(dv, s)
		dependedValues = append(dependedValues, dependedValue)
	})
	value := v.Provide(dependedValues)

	if value.IsError() {
		return value
	}

	if value.IsNil() {
		return errorValue(errors.New("provided `nil` value"))
	}

	err := s.Store(v, value)
	if err != nil {
		return errorValue(err)
	}

	return value
}

func (g *graph) addVertex(key interface{}, v Vertex) Vertex {
	if key != nil {
		g.vertexByKey[key] = v
	}

	return v
}

func (g *graph) AddFuncVertex(f reflect.Value, key interface{}) Vertex {
	return g.addVertex(key, NewFuncVertex(f, -1))
}

func (g *graph) AddIndexedFuncVertex(f reflect.Value, index int, key interface{}) Vertex {
	return g.addVertex(key, NewFuncVertex(f, index))
}

func (g *graph) AddIndexedVertex(index int, key interface{}) Vertex {
	return g.addVertex(key, NewIndexedVertex(index))
}

func (g *graph) AddOneOfVertex(key interface{}) Vertex {
	return g.addVertex(key, NewOneOfVertex())
}

func (g *graph) AddErrorVertex(e error, key interface{}) Vertex {
	return g.addVertex(key, NewErrorVertex(e))
}

func (g *graph) AddArrayVertex(key interface{}) Vertex {
	return g.addVertex(key, NewArrayVertex())
}

func (g *graph) AddValueVertex(v reflect.Value, key interface{}) Vertex {
	return g.addVertex(key, NewValueVertex(v))
}

func (g *graph) AddEdge(from Vertex, to Vertex, order int) {
	g.verticesFromSrc[from] = append(g.verticesFromSrc[from], orderedVertex{to, uint(order)})
	g.verticesToDst[to] = append(g.verticesToDst[to], from)
}

func (g *graph) Validate() error {
	// TODO: cycle detect
	return nil
}
