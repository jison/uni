package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_filteredDigraph_EdgeAttrs(t *testing.T) {
	t.Run("original graph is nil", func(t *testing.T) {
		fg := &filteredDigraph{
			oriGraph: nil,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.EdgeAttrs(1, 2)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("predicate is nil", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{1: 2})

		fg := &filteredDigraph{
			oriGraph:      g,
			nodePredicate: nil,
		}

		attrs, ok := fg.EdgeAttrs(1, 2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{1: 2}, AttrsFrom(attrs))
	})

	t.Run("edge not in filtered graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(2, 3, Attrs{"a": "b"})

		fg := &filteredDigraph{
			oriGraph: g,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.EdgeAttrs(1, 2)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("edge in filtered graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(2, 3, Attrs{"a": "b"})
		fg := &filteredDigraph{
			oriGraph: g,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.EdgeAttrs(2, 3)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, AttrsFrom(attrs))
	})

	t.Run("edge no in the filtered graph and original graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(2, 3, Attrs{"a": "b"})
		fg := &filteredDigraph{
			oriGraph: g,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.EdgeAttrs(3, 4)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})
}

func Test_filteredDigraph_Edges(t *testing.T) {
	t.Run("original graph is nil", func(t *testing.T) {
		fg := &filteredDigraph{
			oriGraph: nil,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(fg.Edges()))
	})

	tests := []struct {
		name      string
		predicate func(Node) bool
		edges     []EdgeEntry
		want      EdgesWithAttrs
	}{
		{"predicate is nil", nil, []EdgeEntry{{1, 2, Attrs{"a": "b"}}},
			EdgesWithAttrs{1: {2: {"a": "b"}}}},
		{"edges has been filtered out",
			func(node Node) bool { return node != 1 },
			[]EdgeEntry{{1, 2, Attrs{"a": "b"}}, {2, 3, Attrs{"c": "d"}}},
			EdgesWithAttrs{2: {3: {"c": "d"}}},
		},
		{"edges has been filtered out 2",
			func(node Node) bool { return node != 2 },
			[]EdgeEntry{
				{1, 2, Attrs{"a": "b"}},
				{2, 3, Attrs{"c": "d"}},
				{3, 4, Attrs{"e": "f"}},
			},
			EdgesWithAttrs{3: {4: {"e": "f"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			for _, e := range tt.edges {
				g.AddEdgeWithAttrs(e.from, e.to, e.attrs)
			}
			fg := &filteredDigraph{
				oriGraph:      g,
				nodePredicate: tt.predicate,
			}
			assert.Equalf(t, tt.want, EdgesWithAttrsFrom(fg.Edges()), "Edges()")
		})
	}
}

func Test_filteredDigraph_InEdgesOf(t *testing.T) {
	t.Run("original graph is nil", func(t *testing.T) {
		fg := &filteredDigraph{
			oriGraph: nil,
			nodePredicate: func(node Node) bool {
				return node != 3
			},
		}

		res := fg.InEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(res))
	})

	tests := []struct {
		name      string
		predicate func(Node) bool
		edges     []EdgeEntry
		node      Node
		want      EdgesWithAttrs
	}{
		{"predicate is nil", nil,
			[]EdgeEntry{{1, 2, Attrs{"a": "b"}}},
			1,
			EdgesWithAttrs{},
		},
		{"edges is filtered out",
			func(node Node) bool { return node != 1 },
			[]EdgeEntry{
				{1, 2, Attrs{"a": "b"}}, {2, 3, Attrs{"c": "d"}},
				{4, 2, Attrs{"c": "d"}},
			},
			2,
			EdgesWithAttrs{4: {2: {"c": "d"}}},
		},
		{"edges is filtered out 2",
			func(node Node) bool { return node != 2 },
			[]EdgeEntry{
				{1, 2, Attrs{"a": "b"}},
				{2, 3, Attrs{"c": "d"}},
				{3, 4, Attrs{"e": "f"}},
			},
			3,
			EdgesWithAttrs{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			for _, e := range tt.edges {
				g.AddEdgeWithAttrs(e.from, e.to, e.attrs)
			}
			fg := &filteredDigraph{
				oriGraph:      g,
				nodePredicate: tt.predicate,
			}
			res := fg.InEdgesOf(tt.node)
			assert.Equalf(t, tt.want, EdgesWithAttrsFrom(res), "InEdgesOf(%v)", tt.node)
		})
	}
}

func Test_filteredDigraph_NodeAttrs(t *testing.T) {
	t.Run("original graph is nil", func(t *testing.T) {
		fg := &filteredDigraph{
			oriGraph: nil,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.NodeAttrs(1)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("predicate is nil", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{1: 2})

		fg := &filteredDigraph{
			oriGraph:      g,
			nodePredicate: nil,
		}

		attrs, ok := fg.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{1: 2}, AttrsFrom(attrs))
	})

	t.Run("node not in filtered graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{"a": "b"})

		fg := &filteredDigraph{
			oriGraph: g,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.NodeAttrs(1)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("node in filtered graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{"a": "b"})
		fg := &filteredDigraph{
			oriGraph: g,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.NodeAttrs(2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, AttrsFrom(attrs))
	})

	t.Run("node no in the filtered graph and original graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{"a": "b"})
		fg := &filteredDigraph{
			oriGraph: g,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		attrs, ok := fg.NodeAttrs(3)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})
}

func Test_filteredDigraph_Nodes(t *testing.T) {
	t.Run("original graph is nil", func(t *testing.T) {
		fg := &filteredDigraph{
			oriGraph: nil,
			nodePredicate: func(node Node) bool {
				return node != 1
			},
		}

		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(fg.Nodes()))
	})

	tests := []struct {
		name      string
		predicate func(Node) bool
		nodes     []NodeEntry
		want      NodesWithAttrs
	}{
		{"predicate is nil", nil, []NodeEntry{{1, Attrs{"a": "b"}}},
			NodesWithAttrs{1: Attrs{"a": "b"}}},
		{"nodes has been filtered out",
			func(node Node) bool { return node != 1 },
			[]NodeEntry{{1, Attrs{"a": "b"}}, {2, Attrs{"c": "d"}}},
			NodesWithAttrs{2: Attrs{"c": "d"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			for _, n := range tt.nodes {
				g.AddNodeWithAttrs(n.node, n.attrs)
			}
			fg := &filteredDigraph{
				oriGraph:      g,
				nodePredicate: tt.predicate,
			}
			assert.Equalf(t, tt.want, NodesWithAttrsFrom(fg.Nodes()), "Nodes()")
		})
	}
}

func Test_filteredDigraph_OutEdgesOf(t *testing.T) {
	t.Run("original graph is nil", func(t *testing.T) {
		fg := &filteredDigraph{
			oriGraph: nil,
			nodePredicate: func(node Node) bool {
				return node != 3
			},
		}

		res := fg.OutEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(res))
	})

	tests := []struct {
		name      string
		predicate func(Node) bool
		edges     []EdgeEntry
		node      Node
		want      EdgesWithAttrs
	}{
		{"predicate is nil", nil,
			[]EdgeEntry{{1, 2, Attrs{"a": "b"}}},
			1,
			EdgesWithAttrs{1: {2: {"a": "b"}}},
		},
		{"edges has been filtered out",
			func(node Node) bool { return node != 4 },
			[]EdgeEntry{
				{1, 2, Attrs{"a": "b"}}, {2, 3, Attrs{"c": "d"}},
				{4, 2, Attrs{"c": "d"}}, {2, 4, Attrs{"e": "f"}},
			},
			2,
			EdgesWithAttrs{2: {3: {"c": "d"}}},
		},
		{"edges has been filtered out 2",
			func(node Node) bool { return node != 2 },
			[]EdgeEntry{
				{1, 2, Attrs{"a": "b"}},
				{2, 3, Attrs{"c": "d"}},
				{3, 4, Attrs{"e": "f"}},
			},
			3,
			EdgesWithAttrs{3: {4: {"e": "f"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			for _, e := range tt.edges {
				g.AddEdgeWithAttrs(e.from, e.to, e.attrs)
			}
			fg := &filteredDigraph{
				oriGraph:      g,
				nodePredicate: tt.predicate,
			}
			res := fg.OutEdgesOf(tt.node)
			assert.Equalf(t, tt.want, EdgesWithAttrsFrom(res), "OutEdgesOf(%v)", tt.node)
		})
	}
}

func Test_filteredEdgeIter_Iterate(t *testing.T) {
	t.Run("original iterator is nil", func(t *testing.T) {
		ei := &filteredEdgeIter{
			oriIter:       nil,
			nodePredicate: nil,
		}
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(ei))
	})

	t.Run("predicate is nil", func(t *testing.T) {
		oriEi := EdgesWithAttrs{1: {2: {}, 3: {"a": "b"}}, 2: {3: {}}}
		ei := &filteredEdgeIter{
			oriIter:       oriEi,
			nodePredicate: nil,
		}
		assert.Equal(t, oriEi, EdgesWithAttrsFrom(ei))
	})

	t.Run("filter non-existent node", func(t *testing.T) {
		oriEi := EdgesWithAttrs{1: {2: {}, 3: {"a": "b"}}, 2: {3: {}}}
		ei := &filteredEdgeIter{
			oriIter: oriEi,
			nodePredicate: func(node Node) bool {
				return node != 4
			},
		}
		expected := oriEi
		assert.Equal(t, expected, EdgesWithAttrsFrom(ei))
	})

	t.Run("filter node", func(t *testing.T) {
		oriEi := EdgesWithAttrs{1: {2: {}, 3: {"a": "b"}}, 2: {3: {}}}
		ei := &filteredEdgeIter{
			oriIter: oriEi,
			nodePredicate: func(node Node) bool {
				return node != 2
			},
		}
		expected := EdgesWithAttrs{1: {3: {"a": "b"}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(ei))
	})

	t.Run("interrupt the iteration", func(t *testing.T) {
		oriEi := EdgesWithAttrs{1: {2: {}, 3: {"a": "b"}}, 2: {3: {}}, 4: {3: {}}}
		ei := &filteredEdgeIter{
			oriIter: oriEi,
			nodePredicate: func(node Node) bool {
				return node != 2
			},
		}

		type key struct {
			from Node
			to   Node
		}
		res := make(map[key]struct{})
		ei.Iterate(func(from Node, to Node, attrs AttrsView) bool {
			k := key{from, to}
			res[k] = struct{}{}
			if len(res) >= 1 {
				return false
			}
			return true
		})

		assert.Equal(t, 1, len(res))
	})
}

func Test_filteredNodeIter_Iterate(t *testing.T) {
	t.Run("original iterator is nil", func(t *testing.T) {
		ni := &filteredNodeIter{
			oriIter:       nil,
			nodePredicate: nil,
		}
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(ni))
	})

	t.Run("predicate is nil", func(t *testing.T) {
		oriNi := NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}, 3: Attrs{"c": "d"}}
		ni := &filteredNodeIter{
			oriIter:       oriNi,
			nodePredicate: nil,
		}

		assert.Equal(t, oriNi, NodesWithAttrsFrom(ni))
	})

	t.Run("filter non-existent node", func(t *testing.T) {
		oriNi := NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}, 3: Attrs{"c": "d"}}
		ni := &filteredNodeIter{
			oriIter: oriNi,
			nodePredicate: func(node Node) bool {
				return node != 4
			},
		}

		assert.Equal(t, oriNi, NodesWithAttrsFrom(ni))
	})

	t.Run("filter node", func(t *testing.T) {
		oriNi := NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}, 3: Attrs{"c": "d"}}
		ni := &filteredNodeIter{
			oriIter: oriNi,
			nodePredicate: func(node Node) bool {
				return node != 3
			},
		}
		expected := NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}}
		assert.Equal(t, expected, NodesWithAttrsFrom(ni))
	})

	t.Run("interrupt the iteration", func(t *testing.T) {
		oriNi := NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}, 3: Attrs{"c": "d"}}
		ni := &filteredNodeIter{
			oriIter: oriNi,
			nodePredicate: func(node Node) bool {
				return node != 2
			},
		}

		res := make(map[Node]struct{})
		ni.Iterate(func(node Node, attrs AttrsView) bool {
			res[node] = struct{}{}

			if len(res) >= 1 {
				return false
			}
			return true
		})

		assert.Equal(t, 1, len(res))
	})
}
