package graph

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func edgesWithAttrsFromMustDistinct(ei EdgeAndAttrsIterator) EdgesWithAttrs {
	edges := EdgesWithAttrs{}
	if ei == nil {
		return edges
	}
	ei.Iterate(func(from Node, to Node, attrs AttrsView) bool {
		var toNodes map[Node]Attrs
		var ok bool
		if toNodes, ok = edges[from]; !ok {
			toNodes = map[Node]Attrs{}
			edges[from] = toNodes
		}

		var edgeAttrs Attrs
		if edgeAttrs, ok = toNodes[to]; !ok {
			edgeAttrs = Attrs{}
			toNodes[to] = edgeAttrs
		} else {
			panic(fmt.Sprintf("edge from %v to %v has already met", from, to))
		}

		attrs.Iterate(func(key interface{}, value interface{}) bool {
			edgeAttrs.Set(key, value)
			return true
		})
		return true
	})

	return edges
}

type _duplicateEdges struct {
	ei EdgeAndAttrsIterator
}

func (de *_duplicateEdges) Iterate(f func(from Node, to Node, attrs AttrsView) bool) bool {
	de.ei.Iterate(f)
	return de.ei.Iterate(f)
}

func Test_edgesWithAttrsFromMustDistinct(t *testing.T) {
	t.Run("edges all distinct", func(t *testing.T) {
		edges := EdgesWithAttrs{
			1: {2: {}, 3: {}},
			3: {4: {}, 5: {}},
		}
		assert.Equal(t, edges, edgesWithAttrsFromMustDistinct(edges))
	})
	t.Run("edges with duplicate edges", func(t *testing.T) {
		assert.Panics(t, func() {
			edges := &_duplicateEdges{
				ei: EdgesWithAttrs{
					1: {2: {}, 3: {}},
					3: {4: {}, 5: {}},
				},
			}
			edgesWithAttrsFromMustDistinct(edges)
		})
	})
}

func TestEdgesWithAttrs_Iterate(t *testing.T) {
	tests := []struct {
		name string
		ei   EdgesWithAttrs
		want EdgesWithAttrs
	}{
		{"edges is nil", nil, EdgesWithAttrs{}},
		{"edges is empty", EdgesWithAttrs{}, EdgesWithAttrs{}},
		{"edges non empty", EdgesWithAttrs{1: {2: Attrs{"a": "b"}, 3: {2: Attrs{"c": "d"}}}},
			EdgesWithAttrs{1: {2: Attrs{"a": "b"}, 3: {2: Attrs{"c": "d"}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := EdgesWithAttrs{}
			tt.ei.Iterate(func(from Node, to Node, attrs AttrsView) bool {
				if _, ok := res[from]; !ok {
					res[from] = map[Node]Attrs{}
				}
				res[from][to] = AttrsFrom(attrs)
				return true
			})

			assert.Equal(t, tt.want, res)
		})
	}

	t.Run("interrupt during iteration", func(t *testing.T) {
		type key struct {
			from Node
			to   Node
		}
		res := make(map[key]struct{})
		ei := EdgesWithAttrs{1: {2: {}, 3: {}}, 2: {3: {}}, 3: {4: {}}}
		didNotInterrupted := ei.Iterate(func(from Node, to Node, attrs AttrsView) bool {
			k := key{from, to}
			res[k] = struct{}{}

			if len(res) >= 2 {
				return false
			}

			return true
		})

		assert.False(t, didNotInterrupted)
		assert.Equal(t, 2, len(res))
	})
}

func Test_edgeIterator_Iterate(t *testing.T) {
	type fields struct {
		edges           EdgesWithAttrs
		reversed        bool
		restrictedNodes []Node
	}
	tests := []struct {
		name   string
		fields fields
		want   EdgesWithAttrs
	}{
		{
			"edges is nil",
			fields{
				nil,
				false,
				nil,
			},
			EdgesWithAttrs{},
		},
		{
			"edges is empty",
			fields{
				EdgesWithAttrs{},
				false,
				nil,
			},
			EdgesWithAttrs{},
		},
		{
			"edges non empty",
			fields{
				EdgesWithAttrs{1: {2: Attrs{"a": "b"}, 3: {2: Attrs{"c": "d"}}}},
				false,
				nil,
			},
			EdgesWithAttrs{1: {2: Attrs{"a": "b"}, 3: {2: Attrs{"c": "d"}}}},
		},
		{
			"reverse",
			fields{
				EdgesWithAttrs{1: {2: Attrs{"a": "b"}, 3: Attrs{"c": "d"}}},
				true,
				nil,
			},
			EdgesWithAttrs{2: {1: Attrs{"a": "b"}}, 3: {1: Attrs{"c": "d"}}},
		},
		{
			"reverse 2",
			fields{
				EdgesWithAttrs{1: {2: Attrs{"a": "b"}}, 3: {2: Attrs{"c": "d"}}},
				true,
				nil,
			},
			EdgesWithAttrs{2: {1: Attrs{"a": "b"}, 3: Attrs{"c": "d"}}},
		},
		{
			"restrict nodes",
			fields{
				EdgesWithAttrs{1: {2: Attrs{"a": "b"}}, 3: {2: Attrs{"c": "d"}}},
				false,
				[]Node{1},
			},
			EdgesWithAttrs{1: {2: Attrs{"a": "b"}}},
		},
		{
			"restrict nodes does not exist",
			fields{
				EdgesWithAttrs{1: {2: Attrs{"a": "b"}}, 3: {2: Attrs{"c": "d"}}},
				false,
				[]Node{2},
			},
			EdgesWithAttrs{},
		},
		{
			"restrict nodes and reverse",
			fields{
				EdgesWithAttrs{1: {2: Attrs{"a": "b"}}, 3: {2: Attrs{"c": "d"}}},
				false,
				[]Node{3},
			},
			EdgesWithAttrs{3: {2: Attrs{"c": "d"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var restrictedNodes nodeIterateFunc
			if tt.fields.restrictedNodes != nil {
				restrictedNodes = func(f func(node Node) bool) bool {
					for _, n := range tt.fields.restrictedNodes {
						if !f(n) {
							return false
						}
					}
					return true
				}
			}

			ei := &edgeIterator{
				edges:            tt.fields.edges,
				reverseDirection: tt.fields.reversed,
				restrictedNodes:  restrictedNodes,
			}
			res := EdgesWithAttrsFrom(ei)
			assert.Equal(t, tt.want, res)
		})
	}
}

func Test_edgeIterator_InterruptDuringIteration(t *testing.T) {
	t.Run("interrupt during iteration", func(t *testing.T) {
		type key struct {
			from Node
			to   Node
		}
		res := make(map[key]struct{})
		ei := &edgeIterator{
			edges:            EdgesWithAttrs{1: {2: {}, 3: {}}, 2: {3: {}}, 3: {4: {}}},
			reverseDirection: false,
			restrictedNodes:  nil,
		}
		didNotInterrupted := ei.Iterate(func(from Node, to Node, attrs AttrsView) bool {
			k := key{from, to}
			res[k] = struct{}{}

			if len(res) >= 2 {
				return false
			}

			return true
		})

		assert.False(t, didNotInterrupted)
		assert.Equal(t, 2, len(res))
	})
}

func TestEdgesWithAttrsFrom(t *testing.T) {
	type args struct {
		ei EdgesWithAttrs
	}
	tests := []struct {
		name string
		args args
		want EdgesWithAttrs
	}{
		{"EdgeAndAttrsIterator is nil", args{nil}, EdgesWithAttrs{}},
		{"EdgeAndAttrsIterator is not nil",
			args{EdgesWithAttrs{1: {2: {"a": "b"}, 3: {"a": "b"}}, 2: {1: {"c": "d"}}}},
			EdgesWithAttrs{1: {2: {"a": "b"}, 3: {"a": "b"}}, 2: {1: {"c": "d"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, EdgesWithAttrsFrom(tt.args.ei), "EdgesWithAttrsFrom(%v)", tt.args.ei)
		})
	}
}

func Test_emptyEdgeIterator_Iterate(t *testing.T) {
	assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(emptyEdgeIterator{}))
}
