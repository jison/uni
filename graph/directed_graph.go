package graph

type DirectedGraphView interface {
	NodeAttrs(node Node) (AttrsView, bool)
	Nodes() NodeAndAttrsIterator

	EdgeAttrs(from, to Node) (AttrsView, bool)
	Edges() EdgeAndAttrsIterator

	OutEdgesOf(node Node) EdgeAndAttrsIterator
	InEdgesOf(node Node) EdgeAndAttrsIterator
}

type DirectedGraph interface {
	DirectedGraphView
	AddNodeWithAttrs(node Node, attrs AttrsView)
	RemoveNode(node Node)

	AddEdgeWithAttrs(from Node, to Node, attrs AttrsView)
	RemoveEdge(from Node, to Node)
}

type digraph struct {
	nodes        NodesWithAttrs
	successors   EdgesWithAttrs
	predecessors EdgesWithAttrs
}

func NewDirectedGraph() DirectedGraph {
	g := &digraph{
		nodes:        NodesWithAttrs{},
		successors:   EdgesWithAttrs{},
		predecessors: EdgesWithAttrs{},
	}

	return g
}

func NewDirectedGraphFrom(source DirectedGraphView) DirectedGraph {
	g := NewDirectedGraph()

	source.Nodes().Iterate(func(node Node, attrs AttrsView) bool {
		g.AddNodeWithAttrs(node, attrs)
		return true
	})

	source.Edges().Iterate(func(from Node, to Node, attrs AttrsView) bool {
		g.AddEdgeWithAttrs(from, to, attrs)
		return true
	})

	return g
}

func (g *digraph) AddNodeWithAttrs(node Node, attrs AttrsView) {
	if !canBeNode(node) {
		return
	}

	var oriAttrs Attrs
	var ok bool
	if oriAttrs, ok = g.nodes[node]; !ok {
		oriAttrs = Attrs{}
		g.nodes[node] = oriAttrs
	}

	if attrs == nil {
		return
	}

	attrs.Iterate(func(key interface{}, value interface{}) bool {
		oriAttrs.Set(key, value)
		return true
	})
}

func (g *digraph) RemoveNode(node Node) {
	if !canBeNode(node) {
		return
	}

	for pre := range g.predecessors[node] {
		delete(g.successors[pre], node)
	}
	delete(g.successors, node)

	for suc := range g.successors[node] {
		delete(g.predecessors[suc], node)
	}
	delete(g.predecessors, node)

	delete(g.nodes, node)
}

func (g *digraph) NodeAttrs(node Node) (AttrsView, bool) {
	if !canBeNode(node) {
		return nil, false
	}

	attrs, ok := g.nodes[node]
	return attrs, ok
}

func (g *digraph) Nodes() NodeAndAttrsIterator {
	return g.nodes
}

func (g *digraph) AddEdgeWithAttrs(from Node, to Node, attrs AttrsView) {
	if !canBeNode(from) || !canBeNode(to) {
		return
	}

	_, fromExist := g.NodeAttrs(from)
	if !fromExist {
		g.AddNodeWithAttrs(from, nil)
	}

	_, toExist := g.NodeAttrs(to)
	if !toExist {
		g.AddNodeWithAttrs(to, nil)
	}

	var successors map[Node]Attrs
	var ok bool
	if successors, ok = g.successors[from]; !ok {
		successors = map[Node]Attrs{}
		g.successors[from] = successors
	}

	var predecessors map[Node]Attrs
	if predecessors, ok = g.predecessors[to]; !ok {
		predecessors = map[Node]Attrs{}
		g.predecessors[to] = predecessors
	}

	var edgeAttrs Attrs
	if edgeAttrs, ok = successors[to]; !ok {
		edgeAttrs = Attrs{}
	}
	successors[to] = edgeAttrs
	predecessors[from] = edgeAttrs

	if attrs != nil {
		attrs.Iterate(func(key interface{}, value interface{}) bool {
			edgeAttrs.Set(key, value)
			return true
		})
	}
}

func (g *digraph) EdgeAttrs(from, to Node) (AttrsView, bool) {
	if !canBeNode(from) || !canBeNode(to) {
		return nil, false
	}
	attrs, ok := g.successors[from][to]
	return attrs, ok
}

func (g *digraph) RemoveEdge(from Node, to Node) {
	if !canBeNode(from) || !canBeNode(to) {
		return
	}

	delete(g.successors[from], to)
	delete(g.predecessors[to], from)
}

func (g *digraph) Edges() EdgeAndAttrsIterator {
	return &edgeIterator{
		edges:            g.successors,
		reverseDirection: false,
		restrictedNodes:  nil,
	}
}

func (g *digraph) OutEdgesOf(node Node) EdgeAndAttrsIterator {
	return &edgeIterator{
		edges:            g.successors,
		reverseDirection: false,
		restrictedNodes:  nodeIterateFunc(func(f func(node Node) bool) bool { return f(node) }),
	}
}

func (g *digraph) InEdgesOf(node Node) EdgeAndAttrsIterator {
	return &edgeIterator{
		edges:            g.predecessors,
		reverseDirection: true,
		restrictedNodes:  nodeIterateFunc(func(f func(node Node) bool) bool { return f(node) }),
	}
}
