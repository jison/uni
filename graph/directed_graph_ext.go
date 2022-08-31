package graph

func HasNode(g DirectedGraphView, node Node) bool {
	_, ok := g.NodeAttrs(node)
	return ok
}

func HasEdge(g DirectedGraphView, from Node, to Node) bool {
	_, ok := g.EdgeAttrs(from, to)
	return ok
}

func SuccessorsOf(g DirectedGraphView, node Node) NodeAndAttrsIterator {
	return &graphNodes{
		graph: g,
		nodes: nodeIterateFunc(func(f func(Node) bool) bool {
			edges := g.OutEdgesOf(node)
			if edges == nil {
				return true
			}

			return edges.Iterate(func(from Node, to Node, attrs AttrsView) bool {
				return f(to)
			})
		}),
	}
}

func PredecessorsOf(g DirectedGraphView, node Node) NodeAndAttrsIterator {
	return &graphNodes{
		graph: g,
		nodes: nodeIterateFunc(func(f func(Node) bool) bool {
			edges := g.InEdgesOf(node)
			if edges == nil {
				return true
			}

			return edges.Iterate(func(from Node, to Node, attrs AttrsView) bool {
				return f(from)
			})
		}),
	}
}

func AddNodes(g DirectedGraph, nodes ...Node) {
	for _, n := range nodes {
		g.AddNodeWithAttrs(n, nil)
	}
}

func AddNodesWithAttrs(g DirectedGraph, nodes NodesWithAttrs) {
	for node, attrs := range nodes {
		g.AddNodeWithAttrs(node, attrs)
	}
}

func RemoveNodes(g DirectedGraph, nodes ...Node) {
	for _, n := range nodes {
		g.RemoveNode(n)
	}
}

func AddEdge(g DirectedGraph, from Node, to Node) {
	g.AddEdgeWithAttrs(from, to, nil)
}

func AddEdges(g DirectedGraph, edges [][2]Node) {
	for _, edge := range edges {
		g.AddEdgeWithAttrs(edge[0], edge[1], nil)
	}
}

func AddEdgesWithAttrs(g DirectedGraph, edges EdgesWithAttrs) {
	for from, toNodes := range edges {
		for to, attrs := range toNodes {
			g.AddEdgeWithAttrs(from, to, attrs)
		}
	}
}

func RemoveEdges(g DirectedGraph, edges [][2]Node) {
	for _, e := range edges {
		g.RemoveEdge(e[0], e[1])
	}
}

func SubGraphOf(source DirectedGraphView, predicate func(Node) bool) DirectedGraphView {
	if predicate == nil {
		return source
	}

	return &filteredDigraph{
		oriGraph:      source,
		nodePredicate: predicate,
	}
}

func SubGraphWithNodes(source DirectedGraphView, nodes []Node) DirectedGraphView {
	nodeSet := make(map[Node]struct{})
	for _, node := range nodes {
		if !canBeNode(node) {
			continue
		}
		nodeSet[node] = struct{}{}
	}

	return SubGraphOf(source, func(node Node) bool {
		_, ok := nodeSet[node]
		return ok
	})
}
