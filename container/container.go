package container

import (
	"reflect"

	"github.com/jison/uni/graph"
	"github.com/jison/uni/internal/errors"
	"github.com/jison/uni/module"
)

type container struct {
	matcher module.ComponentMatcher
	graph   graph.Graph
	storage *globalStorage
}

type globalStorage struct {
	values map[graph.Vertex]graph.VertexValue
}

var _ graph.ScopeStorage = &globalStorage{}

func (s *globalStorage) Load(vertex graph.Vertex) (graph.VertexValue, bool) {
	if v, ok := s.values[vertex]; ok {
		return v, true
	} else {
		return nil, false
	}
}

func (s *globalStorage) Store(vertex graph.Vertex, value graph.VertexValue) error {
	s.values[vertex] = value
	return nil
}

func (c *container) ValuesOf(criteria module.Criteria) ([]reflect.Value, error) {
	errs := make([]error, 0)
	vals := make([]reflect.Value, 0)
	c.matcher.Match(criteria, "").Each(func(com module.Component) {
		vertex := c.graph.VertexBy(com)
		vertexVal := c.graph.Provide(vertex, c.storage)
		if vertexVal.IsError() {
			errs = append(errs, vertexVal.Error())
		} else {
			vals = append(vals, vertexVal.Value())
		}
	})

	if len(errs) > 0 {
		return nil, errors.Merge(errs...)
	}

	return vals, nil
}

func (c *container) Invoke(_ interface{}) {
	panic("not implemented") // TODO: Implement
}

func NewContainer(m module.Module, opts ...Option) (Container, error) {
	coms := m.Components()
	var validateErrors []error
	coms.Each(func(c module.Component) {
		err := c.Validate()
		if err != nil {
			validateErrors = append(validateErrors, err)
		}
	})

	if len(validateErrors) > 0 {
		return nil, errors.Merge(validateErrors...)
	}

	matcher, err := module.MatcherOfComponents(coms)
	if err != nil {
		return nil, err
	}

	g, err := graph.GraphFromComponentMatcher(matcher)
	if err != nil {
		return nil, err
	}

	storage := &globalStorage{
		values: make(map[graph.Vertex]graph.VertexValue, 0),
	}

	return &container{matcher: matcher, graph: g, storage: storage}, nil
}
