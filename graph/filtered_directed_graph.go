package graph

type filteredNodeIter struct {
	oriIter       NodeAndAttrsIterator
	nodePredicate func(Node) bool
}

func (ni *filteredNodeIter) Iterate(f func(node Node, attrs AttrsView) bool) bool {
	if ni.oriIter == nil {
		return true
	}

	if ni.nodePredicate == nil {
		ni.oriIter.Iterate(f)
		return true
	}

	return ni.oriIter.Iterate(func(node Node, attrs AttrsView) bool {
		if ni.nodePredicate(node) {
			return f(node, attrs)
		}
		return true
	})
}

type filteredEdgeIter struct {
	oriIter       EdgeAndAttrsIterator
	nodePredicate func(Node) bool
}

func (ei *filteredEdgeIter) Iterate(f func(from Node, to Node, attrs AttrsView) bool) bool {
	if ei.oriIter == nil {
		return true
	}

	if ei.nodePredicate == nil {
		ei.oriIter.Iterate(f)
		return true
	}

	return ei.oriIter.Iterate(func(from Node, to Node, attrs AttrsView) bool {
		if ei.nodePredicate(from) && ei.nodePredicate(to) {
			return f(from, to, attrs)
		}
		return true
	})
}

type filteredDigraph struct {
	oriGraph      DirectedGraphView
	nodePredicate func(Node) bool
}

func (g *filteredDigraph) NodeAttrs(node Node) (AttrsView, bool) {
	if g.oriGraph == nil {
		return nil, false
	}

	if g.nodePredicate != nil && !g.nodePredicate(node) {
		return nil, false
	}

	return g.oriGraph.NodeAttrs(node)
}

func (g *filteredDigraph) Nodes() NodeAndAttrsIterator {
	if g.oriGraph == nil {
		return emptyNodeIterator{}
	}

	return &filteredNodeIter{
		oriIter:       g.oriGraph.Nodes(),
		nodePredicate: g.nodePredicate,
	}
}

func (g *filteredDigraph) EdgeAttrs(from, to Node) (AttrsView, bool) {
	if g.oriGraph == nil {
		return nil, false
	}

	if g.nodePredicate != nil && !(g.nodePredicate(from) && g.nodePredicate(to)) {
		return nil, false
	}

	return g.oriGraph.EdgeAttrs(from, to)
}

func (g *filteredDigraph) Edges() EdgeAndAttrsIterator {
	if g.oriGraph == nil {
		return emptyEdgeIterator{}
	}

	return &filteredEdgeIter{
		oriIter:       g.oriGraph.Edges(),
		nodePredicate: g.nodePredicate,
	}
}

func (g *filteredDigraph) OutEdgesOf(node Node) EdgeAndAttrsIterator {
	if g.oriGraph == nil {
		return emptyEdgeIterator{}
	}

	if g.nodePredicate != nil && !g.nodePredicate(node) {
		return emptyEdgeIterator{}
	}

	oriOutEdges := g.oriGraph.OutEdgesOf(node)
	if oriOutEdges == nil {
		return emptyEdgeIterator{}
	}

	return &filteredEdgeIter{
		oriIter:       oriOutEdges,
		nodePredicate: g.nodePredicate,
	}
}

func (g *filteredDigraph) InEdgesOf(node Node) EdgeAndAttrsIterator {
	if g.oriGraph == nil {
		return emptyEdgeIterator{}
	}

	if g.nodePredicate != nil && !g.nodePredicate(node) {
		return emptyEdgeIterator{}
	}

	oriInEdges := g.oriGraph.InEdgesOf(node)
	if oriInEdges == nil {
		return emptyEdgeIterator{}
	}

	return &filteredEdgeIter{
		oriIter:       oriInEdges,
		nodePredicate: g.nodePredicate,
	}
}
