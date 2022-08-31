package commands

import (
	"github.com/jison/uni/graph"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildGraphvizGraph(t *testing.T) {
	g := graph.NewDirectedGraph()
	graph.AddEdges(g, [][2]graph.Node{
		{2, 1}, {3, 1}, {4, 3}, {5, 3}, {6, 3}, {7, 4}, {8, 5}, {9, 5}, {10, 6}, {11, 6}, {12, 6},
	})

	for _, n := range []graph.Node{1, 3, 5, 7, 9, 11} {
		g.AddNodeWithAttrs(n, graph.Attrs{"a": "b"})
	}

	vizG, err := BuildGraphvizGraph(g)
	assert.Nil(t, err)
	err = vizG.RenderSVGToFile("./testdata/g1.svg")
	assert.Nil(t, err)
}
