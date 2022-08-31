package graph

type flag struct {
	flag *flag
}

var deletedFlag = func() *flag {
	f := &flag{}
	f.flag = f
	return f
}()

type deleteStatus int

const (
	deleteStatusUnknown  deleteStatus = 0
	deleteStatusDeleted  deleteStatus = 1
	deleteStatusRestored deleteStatus = 2
)

func isAttrsDeleted(a AttrsView) bool {
	return getDeleteStatus(a) == deleteStatusDeleted
}

func deleteAttrs(a AttrsView) {
	a.Iterate(func(key interface{}, _ interface{}) bool {
		a.Del(key)
		return true
	})
	a.Set(deletedFlag, deleteStatusDeleted)
}

func restoreAttrs(a AttrsView) {
	a.Set(deletedFlag, deleteStatusRestored)
}

func getDeleteStatus(a AttrsView) deleteStatus {
	s, ok := a.Get(deletedFlag)
	if !ok {
		return deleteStatusUnknown
	}
	if ds, typeOk := s.(deleteStatus); typeOk {
		return ds
	}

	return deleteStatusUnknown
}

type derivedAttrs struct {
	parentAttrs AttrsView
	attrs       AttrsView
}

var _ AttrsView = &derivedAttrs{}

func (a *derivedAttrs) Get(key interface{}) (interface{}, bool) {
	if a == nil || key == nil {
		return nil, false
	}

	if a.attrs != nil {
		ds := getDeleteStatus(a.attrs)
		if ds == deleteStatusDeleted {
			return nil, false
		}

		v, ok := a.attrs.Get(key)
		if ok {
			if v == deletedFlag {
				return nil, false
			} else {
				return v, true
			}
		}

		if ds == deleteStatusRestored {
			return nil, false
		}
	}

	if a.parentAttrs == nil {
		return nil, false
	}

	return a.parentAttrs.Get(key)
}

func (a *derivedAttrs) Set(key interface{}, value interface{}) {
	if a == nil || key == nil {
		return
	}

	if a.attrs == nil {
		return
	}

	ds := getDeleteStatus(a.attrs)
	if ds == deleteStatusDeleted {
		return
	}

	a.attrs.Set(key, value)
}

func (a *derivedAttrs) Del(key interface{}) {
	a.Set(key, deletedFlag)
}

func (a *derivedAttrs) Has(key interface{}) bool {
	_, ok := a.Get(key)
	return ok
}

func (a *derivedAttrs) Iterate(f func(key interface{}, value interface{}) bool) bool {
	if a == nil {
		return true
	}

	if a.attrs != nil {
		ds := getDeleteStatus(a.attrs)
		if ds == deleteStatusDeleted {
			return true
		}

		isContinue := a.attrs.Iterate(func(key interface{}, value interface{}) bool {
			if key == deletedFlag || value == deletedFlag {
				return true
			}

			return f(key, value)
		})

		if !isContinue {
			return false
		}

		if ds == deleteStatusRestored {
			return true
		}
	}

	if a.parentAttrs == nil {
		return true
	}

	return a.parentAttrs.Iterate(func(key interface{}, value interface{}) bool {
		if a.attrs != nil && a.attrs.Has(key) {
			return true
		}
		if !f(key, value) {
			return false
		}
		return true
	})
}

func (a *derivedAttrs) initialized() bool {
	return true
}

type derivedNodeWithAttrs struct {
	derivedGraph DirectedGraph
	parent       NodeAndAttrsIterator
	nodes        NodeAndAttrsIterator
}

func (d derivedNodeWithAttrs) Iterate(f func(Node, AttrsView) bool) bool {
	if d.derivedGraph == nil {
		return true
	}

	allNodes := map[Node]struct{}{}

	if d.parent != nil {
		d.parent.Iterate(func(node Node, _ AttrsView) bool {
			allNodes[node] = struct{}{}
			return true
		})
	}
	if d.nodes != nil {
		d.nodes.Iterate(func(node Node, _ AttrsView) bool {
			allNodes[node] = struct{}{}
			return true
		})
	}

	for node := range allNodes {
		attrsAtNodes, ok := d.derivedGraph.NodeAttrs(node)
		if !ok {
			continue
		}
		if !f(node, attrsAtNodes) {
			return false
		}
	}
	return true
}

type derivedEdgeWithAttrs struct {
	derivedGraph DirectedGraph
	parentEdges  EdgeAndAttrsIterator
	derivedEdges EdgeAndAttrsIterator
}

func (d *derivedEdgeWithAttrs) Iterate(f func(Node, Node, AttrsView) bool) bool {
	if d.derivedGraph == nil {
		return true
	}

	allEdges := map[[2]Node]struct{}{}
	if d.parentEdges != nil {
		d.parentEdges.Iterate(func(from Node, to Node, _ AttrsView) bool {
			allEdges[[2]Node{from, to}] = struct{}{}
			return true
		})
	}

	if d.derivedEdges != nil {
		d.derivedEdges.Iterate(func(from Node, to Node, _ AttrsView) bool {
			allEdges[[2]Node{from, to}] = struct{}{}
			return true
		})
	}

	for nodes := range allEdges {
		attrs, ok := d.derivedGraph.EdgeAttrs(nodes[0], nodes[1])
		if !ok {
			continue
		}
		if !f(nodes[0], nodes[1], attrs) {
			return false
		}
	}

	return true
}

type derivedDirectedGraph struct {
	parentGraph DirectedGraphView

	nodes        NodesWithAttrs
	successors   EdgesWithAttrs
	predecessors EdgesWithAttrs
}

func DeriveDirectedGraph(parent DirectedGraphView) DirectedGraph {
	return &derivedDirectedGraph{
		parentGraph:  parent,
		nodes:        NodesWithAttrs{},
		successors:   EdgesWithAttrs{},
		predecessors: EdgesWithAttrs{},
	}
}

func (g *derivedDirectedGraph) NodeAttrs(node Node) (AttrsView, bool) {
	if g == nil || !canBeNode(node) {
		return nil, false
	}

	attrs, attrsOk := g.nodes[node]
	if attrsOk && isAttrsDeleted(attrs) {
		return nil, false
	}
	if !attrsOk && (g.parentGraph == nil || !HasNode(g.parentGraph, node)) {
		return nil, false
	}

	aProxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
		a, ok := g.nodes[node]
		if !ok {
			if !createIfNeed {
				return nil, false
			}
			a = Attrs{}
			g.nodes[node] = a
		}
		return a, true
	})

	parentAttrs := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
		if g.parentGraph == nil {
			return nil, false
		}
		return g.parentGraph.NodeAttrs(node)
	})

	return &derivedAttrs{parentAttrs, aProxy}, true
}

func (g *derivedDirectedGraph) Nodes() NodeAndAttrsIterator {
	if g == nil {
		return emptyNodeIterator{}
	}

	if g.parentGraph == nil {
		return &derivedNodeWithAttrs{g, nil, g.nodes}
	}

	return &derivedNodeWithAttrs{g, g.parentGraph.Nodes(), g.nodes}
}

func (g *derivedDirectedGraph) EdgeAttrs(from, to Node) (AttrsView, bool) {
	if g == nil || !canBeNode(from) || !canBeNode(to) {
		return nil, false
	}

	attrs, attrsOk := g.successors[from][to]
	if attrsOk && isAttrsDeleted(attrs) {
		return nil, false
	}

	if !attrsOk && (g.parentGraph == nil || !HasEdge(g.parentGraph, from, to)) {
		return nil, false
	}

	aProxy := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
		return g.getEdgeAttrs(from, to, createIfNeed)
	})
	parentAttrs := attrsProxy(func(createIfNeed bool) (AttrsView, bool) {
		if g.parentGraph == nil {
			return nil, false
		}
		return g.parentGraph.EdgeAttrs(from, to)
	})

	return &derivedAttrs{parentAttrs, aProxy}, true
}

func (g *derivedDirectedGraph) Edges() EdgeAndAttrsIterator {
	if g == nil {
		return emptyEdgeIterator{}
	}

	var edges EdgeAndAttrsIterator
	if g.parentGraph == nil {
		edges = nil
	} else {
		edges = g.parentGraph.Edges()
	}

	return &derivedEdgeWithAttrs{g, edges, &edgeIterator{
		edges:            g.successors,
		reverseDirection: false,
		restrictedNodes:  nil,
	}}
}

func (g *derivedDirectedGraph) OutEdgesOf(node Node) EdgeAndAttrsIterator {
	if g == nil {
		return emptyEdgeIterator{}
	}

	var edges EdgeAndAttrsIterator
	if g.parentGraph == nil {
		edges = nil
	} else {
		edges = g.parentGraph.OutEdgesOf(node)
	}

	return &derivedEdgeWithAttrs{g, edges, &edgeIterator{
		edges:            g.successors,
		reverseDirection: false,
		restrictedNodes:  nodeIterateFunc(func(f func(node Node) bool) bool { return f(node) }),
	}}
}

func (g *derivedDirectedGraph) InEdgesOf(node Node) EdgeAndAttrsIterator {
	if g == nil {
		return emptyEdgeIterator{}
	}

	var edges EdgeAndAttrsIterator
	if g.parentGraph == nil {
		edges = nil
	} else {
		edges = g.parentGraph.InEdgesOf(node)
	}

	return &derivedEdgeWithAttrs{g, edges, &edgeIterator{
		edges:            g.predecessors,
		reverseDirection: true,
		restrictedNodes:  nodeIterateFunc(func(f func(node Node) bool) bool { return f(node) }),
	}}
}

func (g *derivedDirectedGraph) AddNodeWithAttrs(node Node, attrs AttrsView) {
	if g == nil || !canBeNode(node) {
		return
	}

	var oriAttrs Attrs
	var ok bool
	if oriAttrs, ok = g.nodes[node]; !ok {
		oriAttrs = Attrs{}
		g.nodes[node] = oriAttrs
	} else if isAttrsDeleted(oriAttrs) {
		restoreAttrs(oriAttrs)
	}

	if attrs == nil {
		return
	}

	attrs.Iterate(func(key interface{}, value interface{}) bool {
		oriAttrs.Set(key, value)
		return true
	})
}

func (g *derivedDirectedGraph) RemoveNode(node Node) {
	if g == nil || !canBeNode(node) {
		return
	}

	g.InEdgesOf(node).Iterate(func(from Node, to Node, attrs AttrsView) bool {
		g.RemoveEdge(from, to)
		return true
	})

	g.OutEdgesOf(node).Iterate(func(from Node, to Node, attrs AttrsView) bool {
		g.RemoveEdge(from, to)
		return true
	})

	if a, ok := g.nodes[node]; !ok {
		a = Attrs{}
		g.nodes[node] = a
		deleteAttrs(a)
	} else {
		deleteAttrs(a)
	}
}

func (g *derivedDirectedGraph) AddEdgeWithAttrs(from Node, to Node, attrs AttrsView) {
	if g == nil || !canBeNode(from) || !canBeNode(to) {
		return
	}

	if !HasNode(g, from) {
		g.AddNodeWithAttrs(from, nil)
	}

	if !HasNode(g, to) {
		g.AddNodeWithAttrs(to, nil)
	}

	edgeAttrs, ok := g.getEdgeAttrs(from, to, true)
	if !ok {
		return
	}
	if isAttrsDeleted(edgeAttrs) {
		restoreAttrs(edgeAttrs)
	}

	if attrs != nil {
		attrs.Iterate(func(key interface{}, value interface{}) bool {
			edgeAttrs.Set(key, value)
			return true
		})
	}
}

func (g *derivedDirectedGraph) getEdgeAttrs(from Node, to Node, createIfNeed bool) (Attrs, bool) {
	var ok bool
	var successors map[Node]Attrs
	if successors, ok = g.successors[from]; !ok {
		if !createIfNeed {
			return nil, false
		}

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
		if !createIfNeed {
			return nil, false
		}

		edgeAttrs = Attrs{}
		successors[to] = edgeAttrs
	}
	if _, ok = predecessors[from]; !ok {
		predecessors[from] = edgeAttrs
	}

	return edgeAttrs, true
}

func (g *derivedDirectedGraph) RemoveEdge(from Node, to Node) {
	if g == nil || !canBeNode(from) || !canBeNode(to) {
		return
	}

	attrs, ok := g.getEdgeAttrs(from, to, true)
	if ok {
		deleteAttrs(attrs)
	}
}
