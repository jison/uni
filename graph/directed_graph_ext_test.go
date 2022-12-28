package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddEdge(t *testing.T) {
	t.Run("edge value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()

		vals := []interface{}{nil, []interface{}{1}, map[interface{}]interface{}{1: 1}, func() {}}
		for _, val := range vals {
			AddEdge(g, val, 1)
			assert.False(t, HasEdge(g, val, 1))
		}
	})

	t.Run("add an edge with a non-existent node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)

		expected := EdgesWithAttrs{1: {2: {}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("add an edge with a exist node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1, 2)
		AddEdge(g, 1, 2)

		expected := EdgesWithAttrs{1: {2: {}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("add an edge has exist in graph", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 1, 3)
		AddEdge(g, 1, 2)

		expected := EdgesWithAttrs{1: {2: {}, 3: {}}}
		assert.Equal(t, expected, EdgesWithAttrsFrom(g.Edges()))
	})
}

func TestAddEdges(t *testing.T) {
	type args struct {
		edges [][2]Node
	}
	tests := []struct {
		name string
		args args
		want EdgesWithAttrs
	}{
		{"edges is nil", args{nil}, EdgesWithAttrs{}},
		{"edges is empty", args{[][2]Node{}}, EdgesWithAttrs{}},
		{"edges is not empty", args{[][2]Node{{1, 2}}}, EdgesWithAttrs{1: {2: {}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			AddEdges(g, tt.args.edges)
			assert.Equal(t, tt.want, EdgesWithAttrsFrom(g.Edges()))
		})
	}
}

func TestAddNodes(t *testing.T) {
	type args struct {
		nodes []Node
	}
	tests := []struct {
		name string
		args args
		want NodesWithAttrs
	}{
		{"empty nodes", args{[]Node{}}, NodesWithAttrs{}},
		{"node is nil", args{[]Node{nil}}, NodesWithAttrs{}},
		{"single node", args{[]Node{1}}, NodesWithAttrs{1: Attrs{}}},
		{"multiple nodes", args{[]Node{1, 2}}, NodesWithAttrs{1: Attrs{}, 2: Attrs{}}},
		{"add nodes several times", args{[]Node{1, 2, 1, 2, 1}},
			NodesWithAttrs{1: Attrs{}, 2: Attrs{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			AddNodes(g, tt.args.nodes...)

			res := NodesWithAttrsFrom(g.Nodes())
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestHasEdge(t *testing.T) {
	t.Run("edge value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasEdge(g, nil, 2))
		assert.False(t, HasEdge(g, []interface{}{1}, 2))
		assert.False(t, HasEdge(g, map[interface{}]interface{}{1: 1}, 2))
		assert.False(t, HasEdge(g, func() {}, 2))
	})

	t.Run("before add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasEdge(g, 1, 2))
	})

	t.Run("after add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		assert.True(t, HasEdge(g, 1, 2))
		assert.False(t, HasEdge(g, 1, 3))
	})

	t.Run("after remove edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 1, 3)
		g.RemoveEdge(1, 2)
		assert.False(t, HasEdge(g, 1, 2))
		assert.True(t, HasEdge(g, 1, 3))
	})

	t.Run("after remove node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 1, 3)
		RemoveNodes(g, 2)
		assert.False(t, HasEdge(g, 1, 2))
		assert.True(t, HasEdge(g, 1, 3))
	})
}

func TestHasNode(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasNode(g, nil))
		assert.False(t, HasNode(g, []interface{}{1}))
		assert.False(t, HasNode(g, map[interface{}]interface{}{1: 1}))
		assert.False(t, HasNode(g, func() {}))
	})

	t.Run("before add node", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasNode(g, 1))
	})

	t.Run("after add node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1, 2)
		assert.True(t, HasNode(g, 1))
		assert.True(t, HasNode(g, 2))
		assert.False(t, HasNode(g, 3))
	})

	t.Run("after remove node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1, 2, 3)
		RemoveNodes(g, 1, 2)
		assert.False(t, HasNode(g, 1))
		assert.False(t, HasNode(g, 2))
		assert.True(t, HasNode(g, 3))
	})
}

func TestPredecessorsOf(t *testing.T) {
	t.Run("InEdgesOf of graph return nil", func(t *testing.T) {
		g := nilDiGraph{}
		predecessors := PredecessorsOf(g, 1)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(predecessors))
	})

	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			PredecessorsOf(g, nil)
			PredecessorsOf(g, []interface{}{1})
			PredecessorsOf(g, map[interface{}]interface{}{1: 1})
			PredecessorsOf(g, func() {})
		})
	})

	t.Run("node does not exist", func(t *testing.T) {
		g := NewDirectedGraph()
		predecessors := PredecessorsOf(g, 1)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(predecessors))
	})

	t.Run("node without predecessor", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 1, 3)
		predecessors := PredecessorsOf(g, 1)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(predecessors))
	})

	t.Run("node with one predecessor", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		predecessors := PredecessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{1: Attrs{}}, NodesWithAttrsFrom(predecessors))
	})

	t.Run("node with several predecessors", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 2)
		predecessors := PredecessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{1: Attrs{}, 2: Attrs{}, 3: Attrs{}}, NodesWithAttrsFrom(predecessors))
	})

	t.Run("get predecessors of a removed node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 2)
		RemoveNodes(g, 2)
		predecessors := PredecessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(predecessors))
	})

	t.Run("get predecessors which has attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 2)
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(3, Attrs{"c": "d"})
		predecessors := PredecessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}, 3: Attrs{"c": "d"}},
			NodesWithAttrsFrom(predecessors))
	})
}

func TestRemoveEdges(t *testing.T) {
	t.Run("remove nil", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}})
		RemoveEdges(g, nil)

		assert.Equal(t, EdgesWithAttrs{1: {2: {}}, 2: {3: {}}}, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("remove edge does not exist", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasEdge(g, 1, 2))
		RemoveEdges(g, [][2]Node{{1, 2}, {3, 4}})
		assert.False(t, HasEdge(g, 1, 2))
	})

	t.Run("remove existing edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}})
		RemoveEdges(g, [][2]Node{{1, 2}, {3, 4}})
		assert.Equal(t, EdgesWithAttrs{2: {3: {}}}, EdgesWithAttrsFrom(g.Edges()))
	})

	t.Run("remove removed edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}})
		RemoveEdges(g, [][2]Node{{1, 2}, {3, 4}})
		RemoveEdges(g, [][2]Node{{1, 2}, {3, 4}})
		assert.Equal(t, EdgesWithAttrs{2: {3: {}}}, EdgesWithAttrsFrom(g.Edges()))
	})
}

func TestRemoveNodes(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			RemoveNodes(g, nil)
			RemoveNodes(g, []interface{}{1})
			RemoveNodes(g, map[interface{}]interface{}{1: 1})
			RemoveNodes(g, func() {})
		})
	})

	t.Run("before add node", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.False(t, HasNode(g, 1))
		RemoveNodes(g, 1)
		assert.False(t, HasNode(g, 1))
	})

	t.Run("after add node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddNodes(g, 1)
		assert.True(t, HasNode(g, 1))
		RemoveNodes(g, 1)
		assert.False(t, HasNode(g, 1))
	})

	t.Run("after add edge", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		RemoveNodes(g, 2)
		assert.True(t, HasNode(g, 1))
		assert.True(t, HasNode(g, 3))
		assert.False(t, HasEdge(g, 1, 2))
		assert.False(t, HasEdge(g, 2, 3))
	})
}

func TestSubGraphOf(t *testing.T) {
	t.Run("predicate is nil", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}, {2, 4}, {3, 4}})

		sg := SubGraphOf(g, nil)

		assert.Equal(t, NodesWithAttrsFrom(g.Nodes()), NodesWithAttrsFrom(sg.Nodes()))
		assert.Equal(t, EdgesWithAttrsFrom(g.Edges()), EdgesWithAttrsFrom(sg.Edges()))
	})

	t.Run("sub graph", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}, {2, 4}, {3, 4}})

		sg := SubGraphOf(g, func(node Node) bool {
			return node != 2
		})

		assert.Equal(t, NodesWithAttrs{1: Attrs{}, 3: Attrs{}, 4: Attrs{}}, NodesWithAttrsFrom(sg.Nodes()))
		assert.Equal(t, EdgesWithAttrs{3: {4: {}}}, EdgesWithAttrsFrom(sg.Edges()))
	})
}

func TestSubGraphWithNodes(t *testing.T) {
	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}, {2, 4}, {3, 4}})

		sg := NewDirectedGraphWithInNodes(g, []Node{
			1, 3, 4,
			nil, []interface{}{1}, map[interface{}]interface{}{1: 1}, func() {}})
		assert.Equal(t, NodesWithAttrs{1: Attrs{}, 3: Attrs{}, 4: Attrs{}}, NodesWithAttrsFrom(sg.Nodes()))
		assert.Equal(t, EdgesWithAttrs{3: {4: {}}}, EdgesWithAttrsFrom(sg.Edges()))
	})

	t.Run("subgraph", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdges(g, [][2]Node{{1, 2}, {2, 3}, {2, 4}, {3, 4}})

		sg := NewDirectedGraphWithInNodes(g, []Node{1, 3, 4})
		assert.Equal(t, NodesWithAttrs{1: Attrs{}, 3: Attrs{}, 4: Attrs{}}, NodesWithAttrsFrom(sg.Nodes()))
		assert.Equal(t, EdgesWithAttrs{3: {4: {}}}, EdgesWithAttrsFrom(sg.Edges()))
	})
}

func TestSuccessorsOf(t *testing.T) {
	t.Run("OutEdgesOf of graph return nil", func(t *testing.T) {
		g := nilDiGraph{}
		successors := SuccessorsOf(g, 1)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(successors))
	})

	t.Run("node value is invalid", func(t *testing.T) {
		g := NewDirectedGraph()
		assert.NotPanics(t, func() {
			SuccessorsOf(g, nil)
			SuccessorsOf(g, []interface{}{1})
			SuccessorsOf(g, map[interface{}]interface{}{1: 1})
			SuccessorsOf(g, func() {})
		})
	})

	t.Run("node does not exist", func(t *testing.T) {
		g := NewDirectedGraph()
		successors := SuccessorsOf(g, 1)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(successors))
	})

	t.Run("node without successor", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 1)
		successors := SuccessorsOf(g, 1)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(successors))
	})

	t.Run("node with one successor", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 2)
		successors := SuccessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{1: Attrs{}}, NodesWithAttrsFrom(successors))
	})

	t.Run("node with several successors", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 2, 2)
		successors := SuccessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{1: Attrs{}, 2: Attrs{}, 3: Attrs{}}, NodesWithAttrsFrom(successors))
	})

	t.Run("get successors of a removed node", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 1, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 2)
		RemoveNodes(g, 2)
		successors := SuccessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(successors))
	})

	t.Run("get successors which has attrs", func(t *testing.T) {
		g := NewDirectedGraph()
		AddEdge(g, 2, 1)
		AddEdge(g, 3, 2)
		AddEdge(g, 2, 3)
		AddEdge(g, 2, 2)
		g.AddNodeWithAttrs(1, Attrs{"a": "b"})
		g.AddNodeWithAttrs(3, Attrs{"c": "d"})
		successors := SuccessorsOf(g, 2)
		assert.Equal(t, NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{}, 3: Attrs{"c": "d"}},
			NodesWithAttrsFrom(successors))
	})
}

func TestAddNodesWithAttrs(t *testing.T) {
	tests := []struct {
		name  string
		nodes NodesWithAttrs
		want  NodesWithAttrs
	}{
		{"nodes is nil", nil, NodesWithAttrs{}},
		{"nodes is not nil", NodesWithAttrs{1: {}, 2: Attrs{"a": "b"}},
			NodesWithAttrs{1: {}, 2: Attrs{"a": "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			AddNodesWithAttrs(g, tt.nodes)
			res := NodesWithAttrsFrom(g.Nodes())
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestAddEdgesWithAttrs(t *testing.T) {
	tests := []struct {
		name  string
		edges EdgesWithAttrs
		want  EdgesWithAttrs
	}{
		{"edges is nil", nil, EdgesWithAttrs{}},
		{"edges is not nil", EdgesWithAttrs{1: {2: {}, 3: {"a": "b"}}},
			EdgesWithAttrs{1: {2: {}, 3: {"a": "b"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			AddEdgesWithAttrs(g, tt.edges)
			res := EdgesWithAttrsFrom(g.Edges())
			assert.Equal(t, tt.want, res)
		})
	}
}
