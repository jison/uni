package commands

import (
	"bytes"
	"fmt"
	"github.com/jison/uni/graph"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/jison/uni/internal/errors"
)

type GraphvizGraph struct {
	viz    *graphviz.Graphviz
	cGraph *cgraph.Graph
}

func (g *GraphvizGraph) Close() error {
	err := g.cGraph.Close()
	if err != nil {
		return err
	}
	return g.viz.Close()
}

func (g *GraphvizGraph) RenderSVGToFile(path string) error {
	return g.viz.RenderFilename(g.cGraph, graphviz.SVG, path)
}

func (g *GraphvizGraph) RenderSVG() (string, error) {
	buf := new(bytes.Buffer)
	err := g.viz.Render(g.cGraph, graphviz.SVG, buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (g *GraphvizGraph) RenderDOT() (string, error) {
	buf := new(bytes.Buffer)
	err := g.viz.Render(g.cGraph, graphviz.XDOT, buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func DrawGraphSVG(g graph.DirectedGraphView, path string) error {
	gg, err := BuildGraphvizGraph(g)
	if err != nil {
		return err
	}
	return gg.RenderSVGToFile(path)
}

func BuildGraphvizGraph(g graph.DirectedGraphView) (*GraphvizGraph, error) {
	viz := graphviz.New()
	vizGraph, graphErr := viz.Graph()
	if graphErr != nil {
		return nil, graphErr
	}

	nodeByVertex := make(map[graph.Node]*cgraph.Node)

	err := errors.Empty()
	i := 0
	g.Nodes().Iterate(func(v graph.Node, attrs graph.AttrsView) bool {
		nodeTitle := fmt.Sprintf("%v", i)
		node, nodeErr := vizGraph.CreateNode(nodeTitle)
		if nodeErr != nil {
			err = err.AddErrors(nodeErr)
			return true
		}

		attrsSb := strings.Builder{}
		attrsSb.WriteString(fmt.Sprintf("%v\n", v))
		attrs.Iterate(func(key interface{}, value interface{}) bool {
			attrsSb.WriteString(fmt.Sprintf("%v: %v\n", key, value))
			return true
		})
		label := attrsSb.String()
		node.SetLabel(label)

		nodeByVertex[v] = node
		i += 1
		return true
	})

	if err.HasError() {
		return nil, err
	}

	g.Nodes().Iterate(func(v graph.Node, _ graph.AttrsView) bool {
		outNode, outNodeOk := nodeByVertex[v]
		if !outNodeOk {
			err = err.AddErrors(errors.Bugf("unknown vertex"))
			return true
		}

		graph.PredecessorsOf(g, v).Iterate(func(iv graph.Node, _ graph.AttrsView) bool {
			inNode, inNodeOk := nodeByVertex[iv]
			if !inNodeOk {
				err = err.AddErrors(errors.Bugf("unknown vertex"))
				return true
			}
			edge, edgeErr := vizGraph.CreateEdge("", inNode, outNode)
			if edgeErr != nil {
				err = err.AddErrors(edgeErr)
			}
			edge.SetDir(cgraph.ForwardDir)
			return true
		})

		return true
	})

	if err.HasError() {
		return nil, err
	}
	return &GraphvizGraph{viz: viz, cGraph: vizGraph}, nil
}
