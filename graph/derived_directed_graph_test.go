package graph

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_derivedAttrs_Del(t *testing.T) {
	type args struct {
		key interface{}
	}
	tests := []struct {
		name  string
		attrs *derivedAttrs
		args  args
		want  Attrs
	}{
		{"is nil", nil, args{1}, Attrs{}},
		{"attrs is nil", &derivedAttrs{nil, nil}, args{1}, Attrs{}},
		{"attrs is empty", &derivedAttrs{nil, Attrs{}}, args{1}, Attrs{}},
		{"attrs has values", &derivedAttrs{nil, Attrs{1: 2}}, args{"a"},
			Attrs{1: 2}},
		{"del key at attrs", &derivedAttrs{nil, Attrs{1: 2}}, args{1}, Attrs{}},
		{"del nonexistent key", &derivedAttrs{nil, Attrs{1: 2}}, args{3},
			Attrs{1: 2}},
		{"del nil key", &derivedAttrs{nil, Attrs{1: 2}}, args{nil},
			Attrs{1: 2}},
		{"del value in parentEdges", &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}}, args{3},
			Attrs{1: 2}},
		{"del value not in parentEdges", &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}}, args{1},
			Attrs{3: 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oldParentAttrs AttrsView
			var newParentAttrs AttrsView
			if tt.attrs != nil {
				oldParentAttrs = attrsFromMustDistinct(tt.attrs.parentAttrs)
			}

			tt.attrs.Del(tt.args.key)
			assert.Equal(t, tt.want, attrsFromMustDistinct(tt.attrs))

			if tt.attrs != nil {
				newParentAttrs = attrsFromMustDistinct(tt.attrs.parentAttrs)
				assert.Equal(t, oldParentAttrs, newParentAttrs)
			}
		})
	}
}

func Test_derivedAttrs_Get(t *testing.T) {
	tests := []struct {
		name  string
		attrs *derivedAttrs
		key   interface{}
		want  interface{}
		want1 bool
	}{
		{"nil", nil, 1, nil, false},
		{"attrs is nil", &derivedAttrs{nil, nil},
			1, nil, false},
		{"key exists at attrs", &derivedAttrs{nil, Attrs{1: 2, "a": "b"}},
			1, 2, true},
		{"key exists at parent", &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}},
			3, 4, true},
		{"key does not exist", &derivedAttrs{Attrs{3: 4}, Attrs{1: 2, "a": "b"}},
			"aa", nil, false},
		{"have same key with parent", &derivedAttrs{Attrs{1: 4}, Attrs{1: 2, "a": "b"}},
			1, 2, true},
		{"key has been deleted", &derivedAttrs{Attrs{1: 4}, Attrs{1: deletedFlag, "a": "b"}},
			1, nil, false},
		{"attrs have been deleted", &derivedAttrs{Attrs{1: 4},
			Attrs{deletedFlag: deleteStatusDeleted}},
			1, nil, false},
		{"attrs have been restored", &derivedAttrs{Attrs{1: 4},
			Attrs{deletedFlag: deleteStatusRestored, "a": "b"}},
			1, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.attrs.Get(tt.key)
			assert.Equalf(t, tt.want, got, "Get(%v)", tt.key)
			assert.Equalf(t, tt.want1, got1, "Get(%v)", tt.key)
		})
	}

	t.Run("after deleted key in attrs", func(t *testing.T) {
		a := &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}}
		a.Del(1)

		v, ok := a.Get(1)
		assert.False(t, ok)
		assert.Equal(t, nil, v)
	})
	t.Run("after deleted key in parentEdges", func(t *testing.T) {
		a := &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}}
		a.Del(3)

		v, ok := a.Get(3)
		assert.False(t, ok)
		assert.Equal(t, nil, v)
	})
}

func Test_derivedAttrs_Has(t *testing.T) {
	tests := []struct {
		name  string
		attrs *derivedAttrs
		key   interface{}
		want  bool
	}{
		{"nil", nil, 1, false},
		{"attrs is nil", &derivedAttrs{nil, nil},
			1, false},
		{"key exists at attrs", &derivedAttrs{nil, Attrs{1: 2, "a": "b"}},
			1, true},
		{"key exists at parentEdges", &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}},
			3, true},
		{"key does not exist", &derivedAttrs{Attrs{3: 4}, Attrs{1: 2, "a": "b"}},
			"aa", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.attrs.Has(tt.key), "Has(%v)", tt.key)
		})
	}

	t.Run("after deleted key in self", func(t *testing.T) {
		a := &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}}
		a.Del(1)

		ok := a.Has(1)
		assert.False(t, ok)
	})
	t.Run("after deleted key in parentEdges", func(t *testing.T) {
		a := &derivedAttrs{Attrs{3: 4}, Attrs{1: 2}}
		a.Del(3)

		ok := a.Has(3)
		assert.False(t, ok)
	})
}

func Test_derivedAttrs_Iterate(t *testing.T) {
	tests := []struct {
		name  string
		attrs *derivedAttrs
		want  Attrs
	}{
		{"nil", nil, Attrs{}},
		{"attrs is nil", &derivedAttrs{nil, nil}, Attrs{}},
		{"attrs is nil 2", &derivedAttrs{Attrs{1: 2}, nil}, Attrs{1: 2}},
		{"attrs has single value", &derivedAttrs{nil, Attrs{1: 2}}, Attrs{1: 2}},
		{"attrs has multiple values", &derivedAttrs{nil, Attrs{1: 2, "a": "b"}},
			Attrs{1: 2, "a": "b"}},
		{"parentEdges is not empty", &derivedAttrs{Attrs{3: 4, "c": "d"}, Attrs{1: 2, "a": "b"}},
			Attrs{1: 2, 3: 4, "a": "b", "c": "d"}},
		{"parentEdges has same key", &derivedAttrs{Attrs{1: 4, "c": "d"}, Attrs{1: 2, "a": "b"}},
			Attrs{1: 2, "a": "b", "c": "d"}},
		{"only parentEdges have value", &derivedAttrs{Attrs{1: 2, "c": "d"}, nil},
			Attrs{1: 2, "c": "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allAttr := Attrs{}
			tt.attrs.Iterate(func(key interface{}, value interface{}) bool {
				assert.False(t, allAttr.Has(key))
				allAttr[key] = value
				assert.True(t, allAttr.Has(key))
				return true
			})
			assert.EqualValues(t, tt.want, allAttr)
		})
	}

	t.Run("interrupt iteration", func(t *testing.T) {
		attrs := &derivedAttrs{Attrs{1: 2, "a": "b", "c": "d"}, Attrs{3: 4, "c": "d2"}}

		tests2 := []struct {
			name  string
			count int
		}{
			{"interrupt at derivedAttrs", 1},
			{"interrupt after derivedAttrs", 2},
			{"interrupt at parent", 3},
			{"interrupt after parent", 4},
		}

		for _, tt := range tests2 {
			t.Run(tt.name, func(t *testing.T) {
				allAttr := Attrs{}
				isContinue := attrs.Iterate(func(key interface{}, value interface{}) bool {
					assert.False(t, allAttr.Has(key))
					allAttr[key] = value
					assert.True(t, allAttr.Has(key))

					return allAttr.Len() < tt.count
				})
				assert.False(t, isContinue)
				assert.Equal(t, tt.count, allAttr.Len())
			})
		}
	})
}

func Test_derivedAttrs_Set(t *testing.T) {
	type args struct {
		key   interface{}
		value interface{}
	}
	tests := []struct {
		name  string
		attrs *derivedAttrs
		args  args
		want  Attrs
	}{
		{"nil", nil, args{1, 2}, Attrs{}},
		{"attrs is nil", &derivedAttrs{nil, nil},
			args{1, 2}, Attrs{}},
		{"attrs is nil 2", &derivedAttrs{Attrs{1: 2}, nil},
			args{2, 4}, Attrs{1: 2}},
		{"attrs is empty", &derivedAttrs{Attrs{1: 2}, Attrs{}},
			args{2, 4}, Attrs{1: 2, 2: 4}},
		{"update value at attrs", &derivedAttrs{Attrs{1: 2}, Attrs{"a": "b"}},
			args{"a", "c"}, Attrs{1: 2, "a": "c"}},
		{"update value at parentEdges", &derivedAttrs{Attrs{1: 2}, Attrs{"a": "b"}},
			args{1, 3}, Attrs{1: 3, "a": "b"}},
		{"update value at parentEdges and attrs", &derivedAttrs{Attrs{1: 2}, Attrs{1: 3}},
			args{1, 4}, Attrs{1: 4}},
		{"set nil key", &derivedAttrs{Attrs{1: 2}, Attrs{"a": "b"}},
			args{nil, 4}, Attrs{1: 2, "a": "b"}},
		{"set nil value", &derivedAttrs{Attrs{1: 2}, Attrs{"a": "b"}},
			args{1, nil}, Attrs{1: nil, "a": "b"}},
		{"set nonexistent key", &derivedAttrs{Attrs{1: 2}, Attrs{"a": "b"}},
			args{2, 4}, Attrs{1: 2, 2: 4, "a": "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oldParentAttrs AttrsView
			var newParentAttrs AttrsView
			if tt.attrs != nil {
				oldParentAttrs = attrsFromMustDistinct(tt.attrs.parentAttrs)
			}

			tt.attrs.Set(tt.args.key, tt.args.value)
			assert.Equal(t, tt.want, attrsFromMustDistinct(tt.attrs))

			if tt.attrs != nil {
				newParentAttrs = attrsFromMustDistinct(tt.attrs.parentAttrs)
				assert.Equal(t, oldParentAttrs, newParentAttrs)
			}
		})
	}
}

func Test_derivedNodeWithAttrs_Iterate(t *testing.T) {
	type fields struct {
		parent NodeAndAttrsIterator
		nodes  NodesWithAttrs
	}
	tests := []struct {
		name   string
		fields fields
		want   NodesWithAttrs
	}{
		{"parent is nil, nodes is nil",
			fields{
				parent: nil,
				nodes:  nil,
			},
			NodesWithAttrs{},
		},
		{"parent is nil, nodes is not nil",
			fields{
				parent: nil,
				nodes:  NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
		},
		{"parent is not nil, nodes is nil",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  nil,
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
		},
		{"parent and nodes do not have same node",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{3: {"a": "b"}, 4: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}, 3: {"a": "b"}, 4: {"a": "b"}},
		},
		{"parent and nodes have same node",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{2: {"a": "b"}, 3: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}, 3: {"a": "b"}},
		},
		{"the same node in parent and nodes have different attrs",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{2: {"a": "c"}, 3: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "c"}, 3: {"a": "b"}},
		},
		{"the same node in parent and nodes have different attrs 2",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{2: {"c": "d"}, 3: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b", "c": "d"}, 3: {"a": "b"}},
		},
		{"delete node not in parent",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{4: {deletedFlag: deleteStatusDeleted}, 3: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}, 3: {"a": "b"}},
		},
		{"delete node in parent",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{2: {deletedFlag: deleteStatusDeleted}, 3: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 3: {"a": "b"}},
		},
		{
			"restore node in derive graph",
			fields{
				parent: NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
				nodes:  NodesWithAttrs{2: {deletedFlag: deleteStatusRestored, "c": "d"}, 3: {"a": "b"}},
			},
			NodesWithAttrs{1: {"a": "b"}, 2: {"c": "d"}, 3: {"a": "b"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := NewDirectedGraph()
			AddNodesWithAttrs(pg, nodesWithAttrsFromMustDistinct(tt.fields.parent))
			dg := DeriveDirectedGraph(pg)
			AddNodesWithAttrs(dg, nodesWithAttrsFromMustDistinct(tt.fields.nodes))
			d := derivedNodeWithAttrs{
				derivedGraph: dg,
				parent:       tt.fields.parent,
				nodes:        tt.fields.nodes,
			}
			assert.Equal(t, tt.want, nodesWithAttrsFromMustDistinct(d))
		})
	}

	t.Run("derivedGraph is nil", func(t *testing.T) {
		d := derivedNodeWithAttrs{
			derivedGraph: nil,
			parent:       NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
			nodes:        NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}},
		}
		assert.Equal(t, NodesWithAttrs{}, nodesWithAttrsFromMustDistinct(d))
	})

	t.Run("interrupt iteration", func(t *testing.T) {
		parent := NodesWithAttrs{1: {"a": "b"}, 2: {"a": "b"}}
		nodes := NodesWithAttrs{3: {"a": "b"}, 4: {"a": "b"}}

		pg := NewDirectedGraph()
		AddNodesWithAttrs(pg, nodesWithAttrsFromMustDistinct(parent))
		dg := DeriveDirectedGraph(pg)
		AddNodesWithAttrs(dg, nodesWithAttrsFromMustDistinct(nodes))
		it := derivedNodeWithAttrs{
			derivedGraph: dg,
			parent:       parent,
			nodes:        nodes,
		}

		res := NodesWithAttrs{}
		it.Iterate(func(node Node, view AttrsView) bool {
			res[node] = attrsFromMustDistinct(view)
			return len(res) < 2
		})

		assert.Equal(t, 2, len(res))
	})
}

func Test_derivedEdgeWithAttrs_Iterate(t *testing.T) {
	type fields struct {
		parent EdgeAndAttrsIterator
		edges  EdgeAndAttrsIterator
	}
	tests := []struct {
		name   string
		fields fields
		want   EdgesWithAttrs
	}{
		{"parentEdges is nil, edges is nil",
			fields{
				parent: nil,
				edges:  nil,
			},
			EdgesWithAttrs{},
		},
		{"parentEdges is nil, edges is not nil",
			fields{
				parent: nil,
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {2: {"a": "b"}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}},
		},
		{"parentEdges is not nil, edges is nil",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}},
				edges:  nil,
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}},
		},
		{"parentEdges and edges do not have same edge",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{2: {1: {"a": "b"}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}, 2: {1: {"a": "b"}}},
		},
		{"parentEdges and edges do have same edge",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {2: {"a": "c"}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "c"}}, 3: {4: {"a": "b"}}},
		},
		{"add edge attrs",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {2: {"c": "d"}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b", "c": "d"}}, 3: {4: {"a": "b"}}},
		},
		{"delete edge not in parentEdges",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{4: {2: {deletedFlag: deleteStatusDeleted}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
		},
		{"delete edge in parentEdges",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {2: {deletedFlag: deleteStatusDeleted}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{3: {4: {"a": "b"}}},
		},
		{"restore edge",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{3: {4: {deletedFlag: deleteStatusRestored, "c": "d"}}},
					reverseDirection: false,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"c": "d"}}},
		},
		{"reverse direction",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {2: {"c": "d"}}},
					reverseDirection: true,
					restrictedNodes:  nil,
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}, 2: {1: {"c": "d"}}, 3: {4: {"a": "b"}}},
		},
		{"restricted nodes",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {3: {"c": "d"}, 4: {}}, 2: {1: {}, 3: {}}},
					reverseDirection: false,
					restrictedNodes: nodeIterateFunc(func(f func(node Node) bool) bool {
						return f(1)
					}),
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}, 3: {"c": "d"}, 4: {}}, 3: {4: {"a": "b"}}},
		},
		{"reverse direction and restricted nodes",
			fields{
				parent: EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}},
				edges: &edgeIterator{
					edges:            EdgesWithAttrs{1: {3: {"c": "d"}, 4: {}}, 2: {1: {}, 3: {}}},
					reverseDirection: true,
					restrictedNodes: nodeIterateFunc(func(f func(node Node) bool) bool {
						return f(1)
					}),
				},
			},
			EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {1: {"c": "d"}, 4: {"a": "b"}}, 4: {1: {}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg := NewDirectedGraph()
			AddEdgesWithAttrs(pg, edgesWithAttrsFromMustDistinct(tt.fields.parent))
			dg := DeriveDirectedGraph(pg)
			AddEdgesWithAttrs(dg, edgesWithAttrsFromMustDistinct(tt.fields.edges))

			d := &derivedEdgeWithAttrs{
				derivedGraph: dg,
				parentEdges:  tt.fields.parent,
				derivedEdges: tt.fields.edges,
			}
			assert.Equal(t, tt.want, edgesWithAttrsFromMustDistinct(d))
		})
	}

	t.Run("derivedGraph is nil", func(t *testing.T) {
		d := &derivedEdgeWithAttrs{
			derivedGraph: nil,
			parentEdges:  EdgesWithAttrs{1: {2: {"a": "b"}}},
			derivedEdges: nil,
		}
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(d))
	})

	t.Run("interrupt iteration", func(t *testing.T) {
		parentEdges := EdgesWithAttrs{1: {2: {"a": "b"}}, 3: {4: {"a": "b"}}}
		derivedEdges := EdgesWithAttrs{2: {1: {"c": "d"}}, 3: {4: {}}}
		pg := NewDirectedGraph()
		AddEdgesWithAttrs(pg, edgesWithAttrsFromMustDistinct(parentEdges))
		dg := DeriveDirectedGraph(pg)
		AddEdgesWithAttrs(dg, edgesWithAttrsFromMustDistinct(derivedEdges))

		it := &derivedEdgeWithAttrs{
			derivedGraph: dg,
			parentEdges:  parentEdges,
			derivedEdges: derivedEdges,
		}

		res := EdgesWithAttrs{}
		it.Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if _, ok := res[from]; !ok {
				res[from] = map[Node]Attrs{}
			}
			res[from][to] = attrsFromMustDistinct(attrs)
			return len(res) < 2
		})

		assert.Equal(t, 2, len(res))
	})
}

func generateDirectedGraph() DirectedGraphView {
	g := NewDirectedGraph()
	for i := 1; i < 13; i++ {
		g.AddNodeWithAttrs(i, Attrs{fmt.Sprintf("k%d", i): fmt.Sprintf("v%d", i)})
	}
	AddEdgesWithAttrs(g, EdgesWithAttrs{
		2:  {1: {"ek2-1": "ev2-1"}},
		3:  {1: {"ek3-1": "ev3-1"}},
		4:  {3: {"ek4-3": "ev4-3"}},
		5:  {3: {"ek5-3": "ev5-3"}},
		6:  {3: {"ek6-3": "ev6-3"}},
		7:  {4: {"ek7-4": "ev7-4"}},
		8:  {5: {"ek8-5": "ev8-5"}},
		9:  {5: {"ek9-5": "ev9-5"}},
		10: {6: {"ek10-6": "ev10-6"}},
		11: {6: {"ek11-6": "ev11-6"}},
		12: {6: {"ek12-6": "ev12-6"}},
	})
	return g
}

func Test_derivedDirectedGraph_NodeAttrs(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		attrs, ok := dg.NodeAttrs(1)
		assert.False(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))
	})

	t.Run("invalid node", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddNodeWithAttrs(1, nil)

		attrs, ok := dg.NodeAttrs(map[int]int{})
		assert.False(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddNodeWithAttrs(1, nil)

		attrs, ok := dg.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))

		attrs2, ok2 := dg.NodeAttrs(2)
		assert.False(t, ok2)
		assert.Nil(t, attrs2)
	})

	t.Run("node in parentEdges graph", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		for i := 1; i < 13; i++ {
			attrs, ok := dg.NodeAttrs(i)
			assert.True(t, ok)
			assert.Equal(t, Attrs{fmt.Sprintf("k%d", i): fmt.Sprintf("v%d", i)}, attrsFromMustDistinct(attrs))
		}
	})

	t.Run("node in the derived graph", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.AddNodeWithAttrs(20, Attrs{"a": "b"})

		attrs, ok := dg.NodeAttrs(20)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("node not in the graph", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.AddNodeWithAttrs(20, Attrs{"a": "b"})

		attrs, ok := dg.NodeAttrs(21)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("override node attrs in parentEdges", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})
		dg.AddNodeWithAttrs(2, Attrs{"k2": "v2", "a": "b"})

		attrs1, ok1 := dg.NodeAttrs(1)
		assert.True(t, ok1)
		assert.Equal(t, Attrs{"k1": "v1", "a": "b"}, attrsFromMustDistinct(attrs1))

		attrs2, ok2 := dg.NodeAttrs(2)
		assert.True(t, ok2)
		assert.Equal(t, Attrs{"k2": "v2", "a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("node has removed in parentEdges graph", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.RemoveNode(1)

		attrs, ok := dg.NodeAttrs(1)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("node has removed in derived graph", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.AddNodeWithAttrs(20, Attrs{"a": "b"})
		dg.RemoveNode(20)

		attrs, ok := dg.NodeAttrs(20)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("remove node in parentEdges graph and add it from derived graph", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.RemoveNode(1)
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})

		attrs, ok := dg.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("nodes being added by adding edge", func(t *testing.T) {
		dg := DeriveDirectedGraph(generateDirectedGraph())
		dg.AddEdgeWithAttrs(1, 20, nil)

		attrs, ok := dg.NodeAttrs(20)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))
	})

	t.Run("update attrs after getting node attrs in parentEdges graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})

		attrs1, _ := dg.NodeAttrs(1)
		attrs2, _ := dg.NodeAttrs(1)
		attrs2.Del("k1")
		attrs2.Set("a", "b")
		attrs3, _ := dg.NodeAttrs(1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs3))
	})

	t.Run("update attrs after getting node attrs in derived graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})

		attrs1, _ := dg.NodeAttrs(20)
		attrs2, _ := dg.NodeAttrs(20)
		attrs2.Del("k20")
		attrs2.Set("a", "b")
		attrs3, _ := dg.NodeAttrs(20)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs3))
	})

	t.Run("update attrs after getting node attrs in both graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})

		attrs1, _ := dg.NodeAttrs(1)
		attrs2, _ := dg.NodeAttrs(1)
		attrs2.Del("k1")
		attrs2.Set("a", "c")
		attrs3, _ := dg.NodeAttrs(1)
		assert.Equal(t, Attrs{"a": "c"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "c"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"a": "c"}, attrsFromMustDistinct(attrs3))
	})

	t.Run("get attrs before and after removing node", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})

		var attrs1, attrs2, attrs20 AttrsView
		var ok1, ok2, ok20 bool
		attrs1, ok1 = dg.NodeAttrs(1)
		attrs2, ok2 = dg.NodeAttrs(2)
		attrs20, ok20 = dg.NodeAttrs(20)

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.True(t, ok20)

		dg.RemoveNode(1)
		dg.RemoveNode(2)
		dg.RemoveNode(20)

		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs20))

		attrs1.Set("a", "c")
		attrs2.Set("a", "c")
		attrs20.Set("a", "c")
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs20))

		attrs1, ok1 = dg.NodeAttrs(1)
		attrs2, ok2 = dg.NodeAttrs(2)
		attrs20, ok20 = dg.NodeAttrs(20)
		assert.False(t, ok1)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs1))
		assert.False(t, ok2)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2))
		assert.False(t, ok20)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs20))

		dg.AddNodeWithAttrs(1, Attrs{"c": "d"})
		dg.AddNodeWithAttrs(2, Attrs{"c": "d"})
		dg.AddNodeWithAttrs(20, Attrs{"c": "d"})

		attrs1, ok1 = dg.NodeAttrs(1)
		attrs2, ok2 = dg.NodeAttrs(2)
		attrs20, ok20 = dg.NodeAttrs(20)

		assert.True(t, ok1)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs1))
		assert.True(t, ok2)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs2))
		assert.True(t, ok20)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs20))
	})
}

func Test_derivedDirectedGraph_Nodes(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		nodes := dg.Nodes()
		assert.Equal(t, NodesWithAttrs{}, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})

		nodes := dg.Nodes()
		assert.Equal(t, NodesWithAttrs{1: Attrs{"a": "b"}}, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("nodes in parentEdges graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("nodes in derive graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(21, Attrs{"a": "b"})

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		expectNodes[21] = Attrs{"a": "b"}
		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("override node attrs in parentEdges", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})
		dg.AddNodeWithAttrs(2, Attrs{"k2": "vv2", "a": "b"})

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		expectNodes[1] = Attrs{"k1": "v1", "a": "b"}
		expectNodes[2] = Attrs{"k2": "vv2", "a": "b"}

		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("node has removed in parentEdges graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.RemoveNode(1)

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		delete(expectNodes, 1)

		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("node has removed in derived graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(21, Attrs{"a": "b"})
		dg.AddNodeWithAttrs(22, Attrs{"a": "b"})
		dg.RemoveNode(21)

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		expectNodes[22] = Attrs{"a": "b"}

		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("remove node in parentEdges graph and add it from derived graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.RemoveNode(1)
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		expectNodes[1] = Attrs{"a": "b"}

		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("nodes being added by adding edge", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, nil)
		dg.AddEdgeWithAttrs(21, 22, nil)

		nodes := dg.Nodes()

		expectNodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		expectNodes[20] = Attrs{}
		expectNodes[21] = Attrs{}
		expectNodes[22] = Attrs{}

		assert.Equal(t, expectNodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("update node attrs in parentEdges graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})

		attrs1, _ := dg.NodeAttrs(1)
		dg.Nodes().Iterate(func(node Node, attrs AttrsView) bool {
			if node == 1 {
				attrs.Del("k1")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.NodeAttrs(1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update node attrs in derived graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})

		attrs1, _ := dg.NodeAttrs(20)
		dg.Nodes().Iterate(func(node Node, attrs AttrsView) bool {
			if node == 20 {
				attrs.Del("k20")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.NodeAttrs(20)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update node attrs in derived graph with overriding attrs during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})

		attrs1, _ := dg.NodeAttrs(1)
		dg.Nodes().Iterate(func(node Node, attrs AttrsView) bool {
			if node == 1 {
				attrs.Del("k1")
				attrs.Set("a", "c")
			}
			return true
		})
		attrs2, _ := dg.NodeAttrs(1)
		assert.Equal(t, Attrs{"a": "c"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "c"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("remove node during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.AddNodeWithAttrs(21, Attrs{"k21": "v21"})
		dg.Nodes().Iterate(func(node Node, attrs AttrsView) bool {
			if node.(int)%2 == 0 {
				dg.RemoveNode(node)
			}
			return true
		})

		nodes := nodesWithAttrsFromMustDistinct(pg.Nodes())
		for n := range nodes {
			if n.(int)%2 == 0 {
				delete(nodes, n)
			}
		}
		nodes[21] = Attrs{"k21": "v21"}

		assert.Equal(t, nodes, nodesWithAttrsFromMustDistinct(dg.Nodes()))
	})
}

func Test_derivedDirectedGraph_EdgeAttrs(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		attrs, ok := dg.EdgeAttrs(1, 2)
		assert.False(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))
	})

	t.Run("invalid node", func(t *testing.T) {
		var dg *derivedDirectedGraph

		attrs, ok := dg.EdgeAttrs(func() {}, 2)
		assert.False(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))

		attrs2, ok2 := dg.EdgeAttrs(1, map[int]int{})
		assert.False(t, ok2)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2))
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		attrs, ok := dg.EdgeAttrs(1, 2)

		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))

		attrs2, ok2 := dg.EdgeAttrs(2, 1)

		assert.False(t, ok2)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2))
	})

	t.Run("edges in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)

		attrs, ok := dg.EdgeAttrs(2, 1)

		assert.True(t, ok)
		assert.Equal(t, Attrs{"ek2-1": "ev2-1"}, attrsFromMustDistinct(attrs))
	})

	t.Run("edges in derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})

		attrs, ok := dg.EdgeAttrs(20, 21)

		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("override edge attrs in parentEdges", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"ek2-1": "ev2-1-2"})
		dg.AddEdgeWithAttrs(3, 1, Attrs{"a": "b"})

		attrs1, ok1 := dg.EdgeAttrs(2, 1)
		assert.True(t, ok1)
		assert.Equal(t, Attrs{"ek2-1": "ev2-1-2"}, attrsFromMustDistinct(attrs1))

		attrs2, ok2 := dg.EdgeAttrs(3, 1)
		assert.True(t, ok2)
		assert.Equal(t, Attrs{"a": "b", "ek3-1": "ev3-1"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("edge connect nodes in parentEdges graph and derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 20, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"c": "d"})

		attrs1, ok1 := dg.EdgeAttrs(2, 20)
		assert.True(t, ok1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))

		attrs2, ok2 := dg.EdgeAttrs(21, 1)
		assert.True(t, ok2)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("edge has removed in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)

		attrs, ok := dg.EdgeAttrs(2, 1)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("edge has removed in derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"a": "b"})
		dg.RemoveEdge(20, 21)

		attrs, ok := dg.EdgeAttrs(20, 21)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("remove edge in parentEdges graph and add it from derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"a": "b"})

		attrs, ok := dg.EdgeAttrs(2, 1)

		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("update edge attrs after getting edge attrs in parentEdges graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(1, 23, Attrs{"ek1-23": "ev1-23"})
		dg.AddEdgeWithAttrs(24, 1, Attrs{"ek24-1": "ev24-1"})

		attrs1, _ := dg.EdgeAttrs(2, 1)
		attrs2, _ := dg.EdgeAttrs(2, 1)
		attrs2.Del("ek2-1")
		attrs2.Set("a", "b")
		attrs3, _ := dg.EdgeAttrs(2, 1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs3))
	})

	t.Run("update edge attrs after getting edge attrs in derived graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(1, 23, Attrs{"ek1-23": "ev1-23"})
		dg.AddEdgeWithAttrs(24, 1, Attrs{"ek24-1": "ev24-1"})

		attrs1, _ := dg.EdgeAttrs(20, 21)
		attrs2, _ := dg.EdgeAttrs(20, 21)
		attrs2.Del("ek20-21")
		attrs2.Set("a", "b")
		attrs3, _ := dg.EdgeAttrs(20, 21)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs3))
	})

	t.Run("update edge attrs after getting edge attrs in both graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(1, 23, Attrs{"ek1-23": "ev1-23"})
		dg.AddEdgeWithAttrs(24, 1, Attrs{"ek24-1": "ev24-1"})

		attrs1, _ := dg.EdgeAttrs(1, 23)
		attrs2, _ := dg.EdgeAttrs(1, 23)
		attrs2.Del("ek1-23")
		attrs2.Set("a", "b")
		attrs3, _ := dg.EdgeAttrs(1, 23)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs3))
	})

	t.Run("update edge attrs after getting edge attrs which overriding attrs", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"ek2-1": "ev2-1-2", "a": "b"})

		attrs1, _ := dg.EdgeAttrs(2, 1)
		attrs2, _ := dg.EdgeAttrs(2, 1)
		attrs2.Del("a")
		attrs2.Set("c", "d")
		attrs2.Set("ek2-1", "ev2-1-3")
		attrs3, _ := dg.EdgeAttrs(2, 1)
		assert.Equal(t, Attrs{"c": "d", "ek2-1": "ev2-1-3"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"c": "d", "ek2-1": "ev2-1-3"}, attrsFromMustDistinct(attrs2))
		assert.Equal(t, Attrs{"c": "d", "ek2-1": "ev2-1-3"}, attrsFromMustDistinct(attrs3))
	})

	//goland:noinspection GoSnakeCaseUsage
	t.Run("get edge attrs before and after removing edge", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		AddEdgesWithAttrs(dg, EdgesWithAttrs{
			20: {21: Attrs{"ek20-21": "ev20-21"}},
			1:  {20: Attrs{"ek1-20": "ev1-20"}},
			21: {1: Attrs{"ek21-1": "ev21-1"}},
		})

		attrs2_1, ok2_1 := dg.EdgeAttrs(2, 1)
		attrs20_21, ok20_21 := dg.EdgeAttrs(20, 21)
		attrs1_20, ok1_20 := dg.EdgeAttrs(1, 20)
		attrs21_1, ok21_1 := dg.EdgeAttrs(21, 1)

		assert.True(t, ok2_1)
		assert.True(t, ok20_21)
		assert.True(t, ok1_20)
		assert.True(t, ok21_1)

		dg.RemoveEdge(2, 1)
		dg.RemoveEdge(20, 21)
		dg.RemoveEdge(1, 20)
		dg.RemoveEdge(21, 1)

		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2_1))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs20_21))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs1_20))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs21_1))

		attrs2_1.Set("a", "c")
		attrs20_21.Set("a", "c")
		attrs1_20.Set("a", "c")
		attrs21_1.Set("a", "c")
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2_1))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs20_21))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs1_20))
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs21_1))

		attrs2_1, ok2_1 = dg.EdgeAttrs(2, 1)
		attrs20_21, ok20_21 = dg.EdgeAttrs(20, 21)
		attrs1_20, ok1_20 = dg.EdgeAttrs(1, 20)
		attrs21_1, ok21_1 = dg.EdgeAttrs(21, 1)
		assert.False(t, ok2_1)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs2_1))
		assert.False(t, ok20_21)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs20_21))
		assert.False(t, ok1_20)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs1_20))
		assert.False(t, ok21_1)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs21_1))

		dg.AddEdgeWithAttrs(2, 1, Attrs{"c": "d"})
		dg.AddEdgeWithAttrs(20, 21, Attrs{"c": "d"})
		dg.AddEdgeWithAttrs(1, 20, Attrs{"c": "d"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"c": "d"})

		attrs2_1, ok2_1 = dg.EdgeAttrs(2, 1)
		attrs20_21, ok20_21 = dg.EdgeAttrs(20, 21)
		attrs1_20, ok1_20 = dg.EdgeAttrs(1, 20)
		attrs21_1, ok21_1 = dg.EdgeAttrs(21, 1)

		assert.True(t, ok2_1)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs2_1))
		assert.True(t, ok20_21)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs20_21))
		assert.True(t, ok1_20)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs1_20))
		assert.True(t, ok21_1)
		assert.Equal(t, Attrs{"c": "d"}, attrsFromMustDistinct(attrs21_1))
	})
}

func Test_derivedDirectedGraph_Edges(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		edges := dg.Edges()
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		edges := dg.Edges()
		assert.Equal(t, EdgesWithAttrs{1: {2: Attrs{"a": "b"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edges in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edges in derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())
		expectedEdges[20] = map[Node]Attrs{21: {"a": "b"}}

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("override edge attrs in parentEdges", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"ek2-1": "ev2-1-2"})
		dg.AddEdgeWithAttrs(3, 1, Attrs{"a": "b"})

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())
		expectedEdges[2][1] = Attrs{"ek2-1": "ev2-1-2"}
		expectedEdges[3][1] = Attrs{"ek3-1": "ev3-1", "a": "b"}

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge connect nodes in parentEdges graph and derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 20, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"c": "d"})

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())
		expectedEdges[2][20] = Attrs{"a": "b"}
		expectedEdges[21] = map[Node]Attrs{1: {"c": "d"}}

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge has removed in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())
		delete(expectedEdges, 2)

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge has removed in derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"a": "b"})
		dg.RemoveEdge(20, 21)

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())
		expectedEdges[21] = map[Node]Attrs{22: {"a": "b"}}

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("remove edge in parentEdges graph and add it from derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"a": "b"})

		edges := dg.Edges()

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.Edges())
		expectedEdges[2][1] = Attrs{"a": "b"}

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("update edge attrs in parentEdges graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(2, 1)
		dg.Edges().Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 2 && to == 1 {
				attrs.Del("ek2-1")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(2, 1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update edge attrs in derived graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(20, 21)
		dg.Edges().Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 20 && to == 21 {
				attrs.Del("ek20-21")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(20, 21)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update edge attrs in both graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(23, 1)
		dg.Edges().Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 23 && to == 1 {
				attrs.Del("ek23-1")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(23, 1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("remove edge during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		dg.Edges().Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 1 || to == 1 {
				dg.RemoveEdge(from, to)
			}
			return true
		})

		edges := edgesWithAttrsFromMustDistinct(pg.Edges())
		delete(edges, 2)
		delete(edges, 3)
		edges[20] = map[Node]Attrs{21: {"ek20-21": "ev20-21"}}
		edges[21] = map[Node]Attrs{22: {"ek21-22": "ev21-22"}}

		assert.Equal(t, edges, edgesWithAttrsFromMustDistinct(dg.Edges()))
	})
}

func Test_derivedDirectedGraph_OutEdgesOf(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		edges := dg.OutEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("invalid node", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		edges := dg.OutEdgesOf(func() {})
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(1, 3, Attrs{"c": "d"})

		edges := dg.OutEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{1: {2: Attrs{"a": "b"}, 3: Attrs{"c": "d"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edges in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)

		edges := dg.OutEdgesOf(2)

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.OutEdgesOf(2))

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edges in derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})

		edges := dg.OutEdgesOf(20)

		assert.Equal(t, EdgesWithAttrs{20: {21: Attrs{"a": "b"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("override edge attrs in parentEdges", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"ek2-1": "ev2-1-2"})
		dg.AddEdgeWithAttrs(3, 1, Attrs{"a": "b"})

		edges := dg.OutEdgesOf(2)

		assert.Equal(t, EdgesWithAttrs{2: {1: Attrs{"ek2-1": "ev2-1-2"}}}, edgesWithAttrsFromMustDistinct(edges))

		edges2 := dg.OutEdgesOf(3)
		assert.Equal(t, EdgesWithAttrs{3: {1: Attrs{"ek3-1": "ev3-1", "a": "b"}}}, edgesWithAttrsFromMustDistinct(edges2))
	})

	t.Run("edge connect nodes in parentEdges graph and derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 20, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"c": "d"})

		edges := dg.OutEdgesOf(2)

		assert.Equal(t, EdgesWithAttrs{2: {1: Attrs{"ek2-1": "ev2-1"}, 20: Attrs{"a": "b"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge has removed in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)

		edges := dg.OutEdgesOf(2)

		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge has removed in derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(2, 3, Attrs{"c": "d"})
		dg.RemoveEdge(20, 21)
		dg.RemoveEdge(2, 3)

		edges := dg.OutEdgesOf(20)
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))

		edges2 := dg.OutEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{2: {1: Attrs{"ek2-1": "ev2-1"}}}, edgesWithAttrsFromMustDistinct(edges2))
	})

	t.Run("remove edge in parentEdges graph and add it from derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"a": "b"})

		edges := dg.OutEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{2: {1: Attrs{"a": "b"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("update edge attrs in parentEdges graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(2, 1)
		dg.OutEdgesOf(2).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 2 && to == 1 {
				attrs.Del("ek2-1")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(2, 1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update edge attrs in derived graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(20, 21)
		dg.OutEdgesOf(20).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 20 && to == 21 {
				attrs.Del("ek20-21")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(20, 21)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update edge attrs in both graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(1, 24)
		dg.OutEdgesOf(1).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 1 && to == 24 {
				attrs.Del("ek1-24")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(1, 24)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("remove edge during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, Attrs{"ek1-20": "ev1-20"})
		dg.AddEdgeWithAttrs(1, 21, Attrs{"ek1-21": "ev1-21"})
		dg.AddEdgeWithAttrs(1, 22, Attrs{"ek1-22": "ev1-22"})

		dg.OutEdgesOf(1).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 1 && to == 20 {
				dg.RemoveEdge(from, to)
			}
			return true
		})

		edges := EdgesWithAttrs{
			1: {21: Attrs{"ek1-21": "ev1-21"}, 22: Attrs{"ek1-22": "ev1-22"}},
		}

		assert.Equal(t, edges, edgesWithAttrsFromMustDistinct(dg.OutEdgesOf(1)))
	})
}

func Test_derivedDirectedGraph_InEdgesOf(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		edges := dg.InEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("invalid node", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		edges := dg.InEdgesOf(func() {})
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(3, 2, Attrs{"c": "d"})

		edges := dg.InEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{1: {2: Attrs{"a": "b"}}, 3: {2: Attrs{"c": "d"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edges in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)

		edges := dg.InEdgesOf(1)

		expectedEdges := edgesWithAttrsFromMustDistinct(parentGraph.InEdgesOf(1))

		assert.Equal(t, expectedEdges, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edges in derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})

		edges := dg.InEdgesOf(21)

		assert.Equal(t, EdgesWithAttrs{20: {21: Attrs{"a": "b"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("override edge attrs in parentEdges", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"ek2-1": "ev2-1-2"})
		dg.AddEdgeWithAttrs(3, 1, Attrs{"a": "b"})

		edges := dg.InEdgesOf(1)

		assert.Equal(t, EdgesWithAttrs{2: {1: Attrs{"ek2-1": "ev2-1-2"}}, 3: {1: Attrs{"ek3-1": "ev3-1", "a": "b"}}},
			edgesWithAttrsFromMustDistinct(edges))

	})

	t.Run("edge connect nodes in parentEdges graph and derive graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(2, 20, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"c": "d"})

		edges := dg.InEdgesOf(20)

		assert.Equal(t, EdgesWithAttrs{2: {20: Attrs{"a": "b"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge has removed in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)

		edges := dg.InEdgesOf(1)

		assert.Equal(t, EdgesWithAttrs{3: {1: Attrs{"ek3-1": "ev3-1"}}}, edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("edge has removed in derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"a": "b"})
		dg.AddEdgeWithAttrs(2, 3, Attrs{"c": "d"})
		dg.RemoveEdge(20, 21)
		dg.RemoveEdge(2, 3)

		edges := dg.InEdgesOf(21)
		assert.Equal(t, EdgesWithAttrs{}, edgesWithAttrsFromMustDistinct(edges))

		edges2 := dg.InEdgesOf(3)
		assert.Equal(t, EdgesWithAttrs{
			4: {3: Attrs{"ek4-3": "ev4-3"}},
			5: {3: Attrs{"ek5-3": "ev5-3"}},
			6: {3: Attrs{"ek6-3": "ev6-3"}},
		}, edgesWithAttrsFromMustDistinct(edges2))
	})

	t.Run("remove edge in parentEdges graph and add it from derived graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.RemoveEdge(2, 1)
		dg.AddEdgeWithAttrs(2, 1, Attrs{"a": "b"})

		edges := dg.InEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{2: {1: Attrs{"a": "b"}}, 3: {1: Attrs{"ek3-1": "ev3-1"}}},
			edgesWithAttrsFromMustDistinct(edges))
	})

	t.Run("update edge attrs in parentEdges graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(2, 1)
		dg.InEdgesOf(1).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 2 && to == 1 {
				attrs.Del("ek2-1")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(2, 1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update edge attrs in derived graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(20, 21)
		dg.InEdgesOf(21).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 20 && to == 21 {
				attrs.Del("ek20-21")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(20, 21)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("update edge attrs in both graph during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		dg.AddEdgeWithAttrs(23, 1, Attrs{"ek23-1": "ev23-1"})
		dg.AddEdgeWithAttrs(1, 24, Attrs{"ek1-24": "ev1-24"})

		attrs1, _ := dg.EdgeAttrs(23, 1)
		dg.InEdgesOf(1).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 23 && to == 1 {
				attrs.Del("ek23-1")
				attrs.Set("a", "b")
			}
			return true
		})
		attrs2, _ := dg.EdgeAttrs(23, 1)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs1))
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs2))
	})

	t.Run("remove edge during iteration", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 1, Attrs{"ek20-1": "ev20-1"})

		dg.InEdgesOf(1).Iterate(func(from Node, to Node, attrs AttrsView) bool {
			if from == 2 && to == 1 {
				dg.RemoveEdge(from, to)
			}
			return true
		})

		edges := EdgesWithAttrs{
			20: {1: Attrs{"ek20-1": "ev20-1"}},
			3:  {1: Attrs{"ek3-1": "ev3-1"}},
		}

		assert.Equal(t, edges, edgesWithAttrsFromMustDistinct(dg.InEdgesOf(1)))
	})
}

func Test_derivedDirectedGraph_AddNodeWithAttrs(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		assert.NotPanics(t, func() {
			dg.AddNodeWithAttrs(1, Attrs{"a": "b"})
			dg.AddNodeWithAttrs(func() {}, Attrs{"a": "b"})
		})
	})

	t.Run("invalid node", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		assert.NotPanics(t, func() {
			dg.AddNodeWithAttrs(func() {}, Attrs{"a": "b"})
		})
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddNodeWithAttrs(1, nil)

		attrs, ok := dg.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))

		attrs2, ok2 := dg.NodeAttrs(2)
		assert.False(t, ok2)
		assert.Nil(t, attrs2)
	})

	t.Run("add node with nil attrs", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddNodeWithAttrs(20, nil)

		attrs, ok := dg.NodeAttrs(20)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))
	})

	t.Run("add node not in parentEdges graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddNodeWithAttrs(20, Attrs{"a": "b"})

		attrs, ok := dg.NodeAttrs(20)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("override node attrs", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddNodeWithAttrs(1, Attrs{"k1": "v1-2"})
		dg.AddNodeWithAttrs(2, Attrs{"a": "b"})
		attrs, _ := dg.NodeAttrs(3)
		attrs.Del("k3")

		attrs1, ok1 := dg.NodeAttrs(1)
		assert.True(t, ok1)
		assert.Equal(t, Attrs{"k1": "v1-2"}, attrsFromMustDistinct(attrs1))

		attrs2, ok2 := dg.NodeAttrs(2)
		assert.True(t, ok2)
		assert.Equal(t, Attrs{"k2": "v2", "a": "b"}, attrsFromMustDistinct(attrs2))

		attrs3, ok3 := dg.NodeAttrs(3)
		assert.True(t, ok3)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs3))
	})

	t.Run("add node after removed from graph", func(t *testing.T) {
		parentGraph := generateDirectedGraph()
		dg := DeriveDirectedGraph(parentGraph)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.RemoveNode(1)

		dg.AddNodeWithAttrs(1, Attrs{"a": "b"})
		attrs, ok := dg.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))

	})
}

func Test_derivedDirectedGraph_RemoveNode(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		assert.NotPanics(t, func() {
			dg.RemoveNode(1)
			dg.RemoveNode(func() {})
		})
	})

	t.Run("invalid nodes", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)

		assert.NotPanics(t, func() {
			dg.RemoveNode(func() {})
		})
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddNodeWithAttrs(1, nil)

		dg.RemoveNode(1)

		_, ok := dg.NodeAttrs(1)
		assert.False(t, ok)
	})

	t.Run("remove node nonexistent", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.AddNodeWithAttrs(21, Attrs{"k21": "v21"})

		dg.RemoveNode(30)

		_, ok := dg.NodeAttrs(30)
		assert.False(t, ok)
	})

	t.Run("remove node in parentEdges", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.AddNodeWithAttrs(21, Attrs{"k21": "v21"})

		dg.RemoveNode(1)

		_, ok := dg.NodeAttrs(1)
		assert.False(t, ok)

		_, ok = dg.EdgeAttrs(2, 1)
		assert.False(t, ok)

		_, ok = dg.EdgeAttrs(3, 1)
		assert.False(t, ok)
	})

	t.Run("remove node in derived", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.AddNodeWithAttrs(21, Attrs{"k21": "v21"})
		dg.AddEdgeWithAttrs(20, 1, Attrs{"ek20-1": "ev20-1"})
		dg.AddEdgeWithAttrs(21, 20, Attrs{"ek21-20": "ev21-20"})

		dg.RemoveNode(20)

		_, ok := dg.NodeAttrs(20)
		assert.False(t, ok)

		_, ok = dg.EdgeAttrs(20, 1)
		assert.False(t, ok)

		_, ok = dg.EdgeAttrs(21, 20)
		assert.False(t, ok)
	})

	t.Run("remove node has already removed", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddNodeWithAttrs(20, Attrs{"k20": "v20"})
		dg.AddNodeWithAttrs(21, Attrs{"k21": "v21"})

		dg.RemoveNode(1)
		dg.RemoveNode(1)

		_, ok := dg.NodeAttrs(1)
		assert.False(t, ok)

		_, ok = dg.EdgeAttrs(2, 1)
		assert.False(t, ok)

		_, ok = dg.EdgeAttrs(3, 1)
		assert.False(t, ok)
	})
}

func Test_derivedDirectedGraph_AddEdgeWithAttrs(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		assert.NotPanics(t, func() {
			dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
			dg.AddEdgeWithAttrs(func() {}, map[int]int{}, Attrs{"a": "b"})
		})
	})

	t.Run("invalid node", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		assert.NotPanics(t, func() {
			dg.AddEdgeWithAttrs(func() {}, map[int]int{}, Attrs{"a": "b"})
		})
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		attrs, ok := dg.EdgeAttrs(1, 2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("add edge with nil attrs", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, nil)

		attrs, ok := dg.EdgeAttrs(1, 2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, attrsFromMustDistinct(attrs))
	})

	t.Run("add edge not in parentEdges graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})
		attrs, ok := dg.EdgeAttrs(20, 21)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"ek20-21": "ev20-21"}, attrsFromMustDistinct(attrs))
	})

	t.Run("add edge with overriding attrs", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})
		dg.AddEdgeWithAttrs(21, 22, Attrs{"ek21-22": "ev21-22"})

		dg.AddEdgeWithAttrs(2, 1, Attrs{"ek2-1": "ev2-1-2", "a": "b"})
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21-2", "a": "b"})

		attrs, ok := dg.EdgeAttrs(2, 1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"ek2-1": "ev2-1-2", "a": "b"}, attrsFromMustDistinct(attrs))

		attrs, ok = dg.EdgeAttrs(20, 21)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"ek20-21": "ev20-21-2", "a": "b"}, attrsFromMustDistinct(attrs))
	})

	t.Run("add edge between nodes from parentEdges and derived", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, Attrs{"ek1-20": "ev1-20"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"ek21-1": "ev21-1"})

		attrs, ok := dg.EdgeAttrs(1, 20)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"ek1-20": "ev1-20"}, attrsFromMustDistinct(attrs))

		attrs, ok = dg.EdgeAttrs(21, 1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"ek21-1": "ev21-1"}, attrsFromMustDistinct(attrs))
	})

	t.Run("add edge after removed from graph", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, Attrs{"ek1-20": "ev1-20"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"ek21-1": "ev21-1"})

		dg.RemoveEdge(21, 1)

		dg.AddEdgeWithAttrs(21, 1, Attrs{"a": "b"})

		attrs, ok := dg.EdgeAttrs(21, 1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrsFromMustDistinct(attrs))
	})
}

func Test_derivedDirectedGraph_RemoveEdge(t *testing.T) {
	t.Run("derivedDirectedGraph is nil", func(t *testing.T) {
		var dg *derivedDirectedGraph

		assert.NotPanics(t, func() {
			dg.RemoveEdge(1, 2)
		})
	})

	t.Run("invalid nodes", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		assert.NotPanics(t, func() {
			dg.RemoveEdge(func() {}, map[int]int{})
		})
	})

	t.Run("parentEdges is nil", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		dg.RemoveEdge(1, 2)
		_, ok := dg.EdgeAttrs(1, 2)
		assert.False(t, ok)
	})

	t.Run("remove edge nonexistent", func(t *testing.T) {
		dg := DeriveDirectedGraph(nil)
		dg.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		dg.RemoveEdge(2, 3)
		_, ok := dg.EdgeAttrs(2, 3)
		assert.False(t, ok)
	})

	t.Run("remove edge in parentEdges", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, Attrs{"ek1-20": "ev1-20"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"ek21-1": "ev21-1"})
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})

		dg.RemoveEdge(2, 1)
		_, ok := dg.EdgeAttrs(2, 1)
		assert.False(t, ok)
	})

	t.Run("remove edge in derived", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, Attrs{"ek1-20": "ev1-20"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"ek21-1": "ev21-1"})
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})

		dg.RemoveEdge(20, 21)
		_, ok := dg.EdgeAttrs(20, 21)
		assert.False(t, ok)
	})

	t.Run("remove edge has already removed", func(t *testing.T) {
		pg := generateDirectedGraph()
		dg := DeriveDirectedGraph(pg)
		dg.AddEdgeWithAttrs(1, 20, Attrs{"ek1-20": "ev1-20"})
		dg.AddEdgeWithAttrs(21, 1, Attrs{"ek21-1": "ev21-1"})
		dg.AddEdgeWithAttrs(20, 21, Attrs{"ek20-21": "ev20-21"})

		dg.RemoveEdge(20, 21)
		dg.RemoveEdge(20, 21)
		_, ok := dg.EdgeAttrs(20, 21)
		assert.False(t, ok)
	})
}

func Test_getDeleteStatus(t *testing.T) {
	type args struct {
		a AttrsView
	}
	tests := []struct {
		name string
		args args
		want deleteStatus
	}{
		{"have not flag", args{Attrs{}}, deleteStatusUnknown},
		{"have flag", args{Attrs{deletedFlag: deleteStatusDeleted}}, deleteStatusDeleted},
		{"type of flag is wrong", args{Attrs{deletedFlag: 123}}, deleteStatusUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getDeleteStatus(tt.args.a), "getDeleteStatus(%v)", tt.args.a)
		})
	}
}
