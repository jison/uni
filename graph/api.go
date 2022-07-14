package graph

import "reflect"

type VertexValue interface {
	Value() reflect.Value
	IsNil() bool
	Error() error
	IsError() bool
}

type ScopeStorage interface {
	Load(v Vertex) (VertexValue, bool)
	Store(v Vertex, value VertexValue) error
}

type Graph interface {
	Verteies() VertexList
	VertexBy(key interface{}) Vertex
	VerticesToDst(Vertex) VertexList
	VerticesFromSrc(Vertex) VertexList
	Provide(v Vertex, s ScopeStorage) VertexValue
}

type Vertex interface {
	Provide([]VertexValue) VertexValue
}

type VertexList interface {
	Each(func(Vertex))
}

type GraphBuilder interface {
	Graph
	AddFuncVertex(f reflect.Value, key interface{}) Vertex
	AddIndexedFuncVertex(f reflect.Value, index int, key interface{}) Vertex
	AddIndexedVertex(index int, key interface{}) Vertex
	AddOneOfVertex(key interface{}) Vertex
	AddErrorVertex(err error, key interface{}) Vertex
	AddArrayVertex(key interface{}) Vertex
	AddValueVertex(val reflect.Value, key interface{}) Vertex
	AddEdge(from Vertex, to Vertex, order int)
	Validate() error
}
