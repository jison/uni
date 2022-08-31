package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNodesInDirectionMatch(t *testing.T) {
	emptyGraph := func() DirectedGraph {
		return NewDirectedGraph()
	}
	normalGraph := func() DirectedGraph {
		g := emptyGraph()
		AddEdges(g, [][2]Node{
			{2, 1}, {3, 1}, {4, 3}, {5, 3}, {6, 3}, {7, 4}, {8, 5}, {9, 5}, {10, 6}, {11, 6}, {12, 6},
		})
		return g
	}
	makeNodesMatch := func(g DirectedGraph, nodes ...Node) DirectedGraph {
		for _, n := range nodes {
			g.AddNodeWithAttrs(n, Attrs{"a": "b"})
		}
		return g
	}
	allPredecessorsMatchGraph := func() DirectedGraph {
		return makeNodesMatch(normalGraph(), 2, 3)
	}
	allPredecessorsNotMatchGraph := func() DirectedGraph {
		return makeNodesMatch(normalGraph(), 4, 9, 6, 12)
	}
	partialPredecessorMatchGraph := func() DirectedGraph {
		return makeNodesMatch(normalGraph(), 2, 4, 9, 6, 12)
	}
	matchesInDeepSearchGraph := func() DirectedGraph {
		return makeNodesMatch(normalGraph(), 7, 9, 12)
	}
	haveCyclesGraph := func() DirectedGraph {
		g := normalGraph()
		AddEdges(g, [][2]Node{
			{3, 6}, {3, 7}, {5, 8},
		})
		return makeNodesMatch(g, 7, 8, 9, 12)
	}
	startNodeInCycleGraph := func() DirectedGraph {
		g := normalGraph()
		AddEdges(g, [][2]Node{
			{1, 2}, {1, 4}, {1, 8}, {1, 12},
		})
		return makeNodesMatch(g, 7, 8, 9, 12)
	}
	matchesInCrossNodeGraph := func() DirectedGraph {
		g := normalGraph()
		AddEdges(g, [][2]Node{
			{4, 2}, {10, 5},
		})
		return makeNodesMatch(g, 4, 10)
	}

	type args struct {
		node  Node
		graph DirectedGraphView
	}
	tests := []struct {
		name string
		args args
		want NodesWithAttrs
	}{
		{"node is nil", args{nil, matchesInDeepSearchGraph()}, NodesWithAttrs{}},
		{"graph is empty", args{1, emptyGraph()}, NodesWithAttrs{}},
		{"all predecessors match", args{1, allPredecessorsMatchGraph()},
			NodesWithAttrs{2: Attrs{"a": "b"}, 3: Attrs{"a": "b"}}},
		{"all predecessors not match", args{1, allPredecessorsNotMatchGraph()},
			NodesWithAttrs{4: Attrs{"a": "b"}, 6: Attrs{"a": "b"}, 9: Attrs{"a": "b"}}},
		{"partial predecessor match", args{1, partialPredecessorMatchGraph()},
			NodesWithAttrs{2: Attrs{"a": "b"}, 4: Attrs{"a": "b"}, 6: Attrs{"a": "b"}, 9: Attrs{"a": "b"}}},
		{"matches in deep search", args{1, matchesInDeepSearchGraph()},
			NodesWithAttrs{7: Attrs{"a": "b"}, 9: Attrs{"a": "b"}, 12: Attrs{"a": "b"}}},
		{"graph have cycles", args{1, haveCyclesGraph()},
			NodesWithAttrs{7: Attrs{"a": "b"}, 8: Attrs{"a": "b"}, 9: Attrs{"a": "b"}, 12: Attrs{"a": "b"}}},
		{"start node in cycle", args{1, startNodeInCycleGraph()},
			NodesWithAttrs{7: Attrs{"a": "b"}, 8: Attrs{"a": "b"}, 9: Attrs{"a": "b"}, 12: Attrs{"a": "b"}}},
		{"matches in cross node", args{1, matchesInCrossNodeGraph()},
			NodesWithAttrs{4: Attrs{"a": "b"}, 10: Attrs{"a": "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := GetNodesInDirectionMatch(tt.args.node,
				func(node Node) NodeAndAttrsIterator {
					return PredecessorsOf(tt.args.graph, node)
				},
				func(node Node, view AttrsView) bool {
					return view.Has("a")
				},
			)
			assert.Equal(t, tt.want, NodesWithAttrsFrom(res))
		})
	}
}
