package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDirectedGraph(t *testing.T) {
	t.Run("new directed graph", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotNil(t, g)
	})
}

func TestNewDirectedGraphFrom(t *testing.T) {
	t.Run("new graph equal the original graph", func(t *testing.T) {
		g1 := NewDirectedGraph()
		g1.AddNodeWithAttrs(1, Attrs{1: 2})
		g1.AddNodeWithAttrs("a", Attrs{"a": "b"})
		g1.AddEdgeWithAttrs(1, "a", Attrs{1: "a"})

		g2 := NewDirectedGraphFrom(g1)
		assert.Equal(t, g1.Nodes(), g2.Nodes())
		assert.Equal(t, g1.Edges(), g2.Edges())
	})

	t.Run("update node isolation", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{1: 2})
		g.AddNodeWithAttrs("a", Attrs{"a": "b"})
		g.AddEdgeWithAttrs(1, "a", Attrs{1: "a"})

		g2 := NewDirectedGraphFrom(g)
		g2.AddNodeWithAttrs(2, Attrs{1: 2})

		assert.NotEqual(t, g.Nodes(), g2.Nodes())
		assert.Equal(t, g.Edges(), g2.Edges())

		g3 := NewDirectedGraphFrom(g)
		g3.AddNodeWithAttrs("a", Attrs{"c": "d"})

		assert.NotEqual(t, g.Nodes(), g3.Nodes())
		assert.Equal(t, g.Edges(), g3.Edges())

		g4 := NewDirectedGraphFrom(g)
		g4.AddNodeWithAttrs("a", Attrs{"a": "d"})

		assert.NotEqual(t, g.Nodes(), g4.Nodes())
		assert.Equal(t, g.Edges(), g4.Edges())
	})

	t.Run("update edge isolation", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{1: 2})
		g.AddNodeWithAttrs("a", Attrs{"a": "b"})
		g.AddEdgeWithAttrs(1, "a", Attrs{1: "a"})

		g2 := NewDirectedGraphFrom(g)
		g2.AddEdgeWithAttrs(2, 3, nil)
		assert.NotEqual(t, g.Nodes(), g2.Nodes())
		assert.NotEqual(t, g.Edges(), g2.Edges())

		g3 := NewDirectedGraphFrom(g)
		g3.AddEdgeWithAttrs(1, 2, nil)
		assert.NotEqual(t, g.Nodes(), g3.Nodes())
		assert.NotEqual(t, g.Edges(), g3.Edges())

		g4 := NewDirectedGraphFrom(g)
		g4.AddEdgeWithAttrs(1, "a", nil)
		assert.Equal(t, g.Nodes(), g4.Nodes())
		assert.Equal(t, g.Edges(), g4.Edges())

		g5 := NewDirectedGraphFrom(g)
		g5.AddEdgeWithAttrs(1, "a", Attrs{1: "b"})
		assert.Equal(t, g.Nodes(), g5.Nodes())
		assert.NotEqual(t, g.Edges(), g5.Edges())
	})
}

func Test_digraph_AddEdgeWithAttrs(t *testing.T) {
	t.Run("edge value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()

		vals := []interface{}{nil, []interface{}{1}, map[interface{}]interface{}{1: 1}, func() {}}
		for _, val := range vals {
			g.AddEdgeWithAttrs(val, 1, nil)
			assert.False(t, HasEdge(g, val, 1))
		}
	})
	t.Run("add an edge with a non-existent node", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, nil)

		expected := EdgesWithAttrs{1: {2: {}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("add an edge with a exist node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1, 2)
		g.AddEdgeWithAttrs(1, 2, nil)

		expected := EdgesWithAttrs{1: {2: {}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("add an edge has exist in graph", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, nil)
		g.AddEdgeWithAttrs(1, 3, nil)
		g.AddEdgeWithAttrs(1, 2, nil)

		expected := EdgesWithAttrs{1: {2: {}, 3: {}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("update with not nil attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		expected := EdgesWithAttrs{1: {2: {"a": "b"}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("update with additional attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(1, 2, Attrs{"c": "d"})

		expected := EdgesWithAttrs{1: {2: {"a": "b", "c": "d"}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("override attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "c"})

		expected := EdgesWithAttrs{1: {2: {"a": "c"}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})
}

func Test_digraph_AddNodeWithAttrs(t *testing.T) {
	type args struct {
		node  Node
		attrs Attrs
	}
	tests := []struct {
		name string
		args args
		want Attrs
	}{
		{"nil attrs", args{1, nil}, Attrs{}},
		{"single attrs", args{1, Attrs{"a": "b"}}, Attrs{"a": "b"}},
		{"multiple attrs", args{1, Attrs{"a": "b", 1: 2}}, Attrs{"a": "b", 1: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			g.AddNodeWithAttrs(tt.args.node, tt.args.attrs)
			attrs, ok := g.NodeAttrs(tt.args.node)
			assert.True(t, ok)
			assert.Equal(t, tt.want, AttrsFrom(attrs))
		})
	}

	t.Run("add additional attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(1, Attrs{1: 2})
		attrs, ok := g.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b", 1: 2}, AttrsFrom(attrs))
	})

	t.Run("override attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b", 1: 2})
		g.AddNodeWithAttrs(1, Attrs{"a": "c"})
		attrs, ok := g.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "c", 1: 2}, AttrsFrom(attrs))
	})

	t.Run("invalid values", func(t *testing.T) {
		g := NewDirectedGraph()

		vals := []interface{}{nil, []interface{}{1}, map[interface{}]interface{}{1: 1}, func() {}}
		for _, val := range vals {
			g.AddNodeWithAttrs(val, Attrs{"a": "b"})
			attrs, ok := g.NodeAttrs(val)
			assert.False(t, ok)
			assert.Nil(t, attrs)
		}
	})
}

func Test_digraph_EdgeAttrs(t *testing.T) {
	t.Run("before add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		attrs, ok := g.EdgeAttrs(1, 2)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("after add edge with nil attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, nil)
		attrs, ok := g.EdgeAttrs(1, 2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, attrs)

		g.AddNodeWithAttrs(3, Attrs{"a": "b"})
		attrs, ok = g.NodeAttrs(3)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, attrs)
	})

	t.Run("after add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		attrs, ok := g.EdgeAttrs(1, 2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, AttrsFrom(attrs))
	})

	t.Run("after remove edge", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(1, 3, Attrs{1: 2})
		g.RemoveEdge(1, 2)

		attrs, ok := g.EdgeAttrs(1, 2)
		assert.False(t, ok)
		assert.Nil(t, attrs)

		attrs, ok = g.EdgeAttrs(1, 3)
		assert.True(t, ok)
		assert.Equal(t, Attrs{1: 2}, AttrsFrom(attrs))
	})

	t.Run("after update attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(1, 3, Attrs{1: 2})
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "c"})
		attrs, ok := g.EdgeAttrs(1, 2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "c"}, AttrsFrom(attrs))

		attrs, ok = g.EdgeAttrs(1, 3)
		assert.True(t, ok)
		assert.Equal(t, Attrs{1: 2}, AttrsFrom(attrs))
	})
}

func Test_digraph_Edges(t *testing.T) {
	t.Run("before add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotNil(t, g.Edges())
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("after add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		g.AddEdgeWithAttrs(2, 3, Attrs{"a": "b"})
		expected := EdgesWithAttrs{
			1: {2: {}},
			2: {3: {"a": "b"}},
		}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("after remove edge", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(2, 3, Attrs{1: 2})
		g.RemoveEdge(1, 2)

		expected := EdgesWithAttrs{
			2: {3: {1: 2}},
		}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("after remove node", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(2, 3, Attrs{1: 2})
		RemoveNodes(g, 1)

		expected := EdgesWithAttrs{
			2: {3: {1: 2}},
		}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("after update attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})
		g.AddEdgeWithAttrs(2, 3, Attrs{1: 2})
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "c"})

		expected := EdgesWithAttrs{
			1: {2: {"a": "c"}},
			2: {3: {1: 2}},
		}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})
}

func Test_digraph_InEdgesOf(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			g.InEdgesOf(nil)
			g.InEdgesOf([]interface{}{1})
			g.InEdgesOf(map[interface{}]interface{}{1: 1})
			g.InEdgesOf(func() {})
		})
	})

	t.Run("node does not exist", func(t *testing.T) {
		g := NewDirectedGraph()
		inEdges := g.InEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(inEdges))
	})

	t.Run("node without in edges", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 1, 3)
		inEdges := g.InEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(inEdges))
	})

	t.Run("node with one in edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		inEdges := g.InEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{1: {2: {}}}, EdgesWithAttrsFrom(inEdges))
	})

	t.Run("node with several in edges", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 2)
		inEdges := g.InEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{1: {2: {}}, 2: {2: {}}, 3: {2: {}}}, EdgesWithAttrsFrom(inEdges))
	})

	t.Run("get in edges of a removed node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 2)
		RemoveNodes(g, 2)
		inEdges := g.InEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(inEdges))
	})

	t.Run("get out edges with attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		outEdges := g.InEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{1: {2: {"a": "b"}}}, EdgesWithAttrsFrom(outEdges))
	})
}

func Test_digraph_NodeAttrs(t *testing.T) {
	t.Run("before add node", func(t *testing.T) {
		g := NewDirectedGraph()
		attrs, ok := g.NodeAttrs(1)
		assert.False(t, ok)
		assert.Nil(t, attrs)
	})

	t.Run("after add node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1, 2)

		attrs, ok := g.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, AttrsFrom(attrs))

		attrs, ok = g.NodeAttrs(2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{}, AttrsFrom(attrs))

		g.AddNodeWithAttrs(3, Attrs{"a": "b"})

		attrs, ok = g.NodeAttrs(3)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "b"}, AttrsFrom(attrs))
	})

	t.Run("after remove node", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{1: 2})
		RemoveNodes(g, 1)

		attrs, ok := g.NodeAttrs(1)
		assert.False(t, ok)
		assert.Nil(t, attrs)

		attrs, ok = g.NodeAttrs(2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{1: 2}, AttrsFrom(attrs))
	})

	t.Run("after update attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{1: 2})
		g.AddNodeWithAttrs(1, Attrs{"a": "c"})

		attrs, ok := g.NodeAttrs(1)
		assert.True(t, ok)
		assert.Equal(t, Attrs{"a": "c"}, AttrsFrom(attrs))

		attrs, ok = g.NodeAttrs(2)
		assert.True(t, ok)
		assert.Equal(t, Attrs{1: 2}, AttrsFrom(attrs))
	})
}

func Test_digraph_Nodes(t *testing.T) {
	t.Run("before add node", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotNil(t, g.Nodes())
		assert.Equal(t, NodesWithAttrs{}, g.Nodes())
	})

	t.Run("after add node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1, 2)

		g.AddNodeWithAttrs(3, Attrs{"a": "b"})
		expected := NodesWithAttrs{
			1: Attrs{},
			2: Attrs{},
			3: Attrs{"a": "b"},
		}
		assert.Equal(t, expected, g.Nodes())
	})

	t.Run("after remove node", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{1: 2})
		RemoveNodes(g, 1)

		expected := NodesWithAttrs{
			2: Attrs{1: 2},
		}
		assert.Equal(t, expected, g.Nodes())
	})

	t.Run("after update attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(2, Attrs{1: 2})
		g.AddNodeWithAttrs(1, Attrs{"a": "c"})

		expected := NodesWithAttrs{
			1: Attrs{"a": "c"},
			2: Attrs{1: 2},
		}
		assert.Equal(t, expected, g.Nodes())
	})

	t.Run("after add edge with additional node", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}})

		expected := NodesWithAttrs{
			1: Attrs{"a": "b"},
			2: Attrs{},
			3: Attrs{},
		}
		assert.Equal(t, expected, g.Nodes())
	})
}

func Test_digraph_OutEdgesOf(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			g.OutEdgesOf(nil)
			g.OutEdgesOf([]interface{}{1})
			g.OutEdgesOf(map[interface{}]interface{}{1: 1})
			g.OutEdgesOf(func() {})
		})
	})

	t.Run("node does not exist", func(t *testing.T) {
		g := NewDirectedGraph()
		outEdges := g.OutEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(outEdges))
	})

	t.Run("node has no out edges", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 1)
		outEdges := g.OutEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(outEdges))
	})

	t.Run("node has one out edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		outEdges := g.OutEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{2: {3: {}}}, EdgesWithAttrsFrom(outEdges))
	})

	t.Run("node has several out edges", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 2, 2)
		outEdges := g.OutEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{2: {1: {}, 2: {}, 3: {}}}, EdgesWithAttrsFrom(outEdges))
	})

	t.Run("get out edges of a removed node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 2, 2)
		RemoveNodes(g, 2)
		outEdges := g.OutEdgesOf(2)
		assert.Equal(t, EdgesWithAttrs{}, EdgesWithAttrsFrom(outEdges))
	})

	t.Run("get out edges with attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		g.AddEdgeWithAttrs(1, 2, Attrs{"a": "b"})

		outEdges := g.OutEdgesOf(1)
		assert.Equal(t, EdgesWithAttrs{1: {2: {"a": "b"}}}, EdgesWithAttrsFrom(outEdges))
	})
}

func Test_digraph_RemoveEdge(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			g.RemoveEdge(nil, 1)
			g.RemoveEdge([]interface{}{1}, 1)
			g.RemoveEdge(map[interface{}]interface{}{1: 1}, 1)
			g.RemoveEdge(func() {}, 1)
		})
	})

	t.Run("remove edge does not exist", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasEdge(g, 1, 2))
		g.RemoveEdge(1, 2)
		assert.False(t, HasEdge(g, 1, 2))
	})

	t.Run("remove existing edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		assert.True(t, HasEdge(g, 1, 2))
		g.RemoveEdge(1, 2)
		assert.False(t, HasEdge(g, 1, 2))
	})

	t.Run("remove removed edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		g.RemoveEdge(1, 2)
		g.RemoveEdge(1, 2)
		assert.False(t, HasEdge(g, 1, 2))
	})
}

func Test_digraph_RemoveNode(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			g.RemoveNode(nil)
			g.RemoveNode([]interface{}{1})
			g.RemoveNode(map[interface{}]interface{}{1: 1})
			g.RemoveNode(func() {})
		})
	})

	t.Run("before add node", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasNode(g, 1))
		g.RemoveNode(1)
		assert.False(t, HasNode(g, 1))
	})

	t.Run("after add node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1)
		assert.True(t, HasNode(g, 1))
		g.RemoveNode(1)
		assert.False(t, HasNode(g, 1))
	})

	t.Run("after add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		g.RemoveNode(2)
		assert.True(t, HasNode(g, 1))
		assert.True(t, HasNode(g, 3))
		assert.False(t, HasEdge(g, 1, 2))
		assert.False(t, HasEdge(g, 2, 3))
	})
}
