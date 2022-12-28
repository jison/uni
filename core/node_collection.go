package core

type NodeIterator interface {
	Iterate(func(Node) bool) bool
}

type NodeSlice []Node

type NodeCollection interface {
	NodeIterator
	Each(func(Node))
	Filter(func(Node) bool) NodeCollection
	ToArray() NodeSlice
	ToSet() NodeSet
}

type filteredNodeIterator struct {
	original  NodeIterator
	predicate func(Node) bool
}

func (ni *filteredNodeIterator) Iterate(f func(Node) bool) bool {
	if ni.original == nil {
		return true
	}

	if ni.predicate == nil {
		return ni.original.Iterate(f)
	}

	return ni.original.Iterate(func(node Node) bool {
		if ni.predicate(node) {
			return f(node)
		}
		return true
	})
}

func _nodesEach(ni NodeIterator, f func(Node)) {
	ni.Iterate(func(node Node) bool {
		f(node)
		return true
	})
}

func _nodesFilter(ni NodeIterator, p func(Node) bool) NodeCollection {
	return &nodeCollection{&filteredNodeIterator{ni, p}}
}

func _nodesToArray(ni NodeIterator) []Node {
	var nodes []Node
	ni.Iterate(func(node Node) bool {
		nodes = append(nodes, node)
		return true
	})
	return nodes
}

func _nodesToSet(ni NodeIterator) NodeSet {
	set := newNodeSet()
	ni.Iterate(func(node Node) bool {
		set.Add(node)
		return true
	})
	return set
}

func (ns NodeSlice) Iterate(f func(Node) bool) bool {
	for _, n := range ns {
		if !f(n) {
			return false
		}
	}
	return true
}

func (ns NodeSlice) Each(f func(Node)) {
	_nodesEach(ns, f)
}

func (ns NodeSlice) Filter(f func(Node) bool) NodeCollection {
	return _nodesFilter(ns, f)
}

func (ns NodeSlice) ToArray() NodeSlice {
	return _nodesToArray(ns)
}

func (ns NodeSlice) ToSet() NodeSet {
	return _nodesToSet(ns)
}

func NewNodeCollection(ni NodeIterator) NodeCollection {
	return &nodeCollection{ni}
}

type nodeCollection struct {
	ni NodeIterator
}

func (nc *nodeCollection) Iterate(f func(Node) bool) bool {
	if nc.ni == nil {
		return true
	}
	return nc.ni.Iterate(f)
}

func (nc *nodeCollection) Each(f func(Node)) {
	if nc.ni == nil {
		return
	}
	_nodesEach(nc.ni, f)
}

func (nc *nodeCollection) Filter(f func(Node) bool) NodeCollection {
	if nc.ni == nil {
		return NodeSlice{}
	}

	return _nodesFilter(nc.ni, f)
}

func (nc *nodeCollection) ToArray() NodeSlice {
	if nc.ni == nil {
		return NodeSlice{}
	}

	return _nodesToArray(nc.ni)
}

func (nc *nodeCollection) ToSet() NodeSet {
	if nc.ni == nil {
		return nodeSet{}
	}

	return _nodesToSet(nc.ni)
}

func newNodeSet(nodes ...Node) nodeSet {
	set := nodeSet{}
	for _, node := range nodes {
		set.Add(node)
	}
	return set
}

type NodeSet interface {
	NodeCollection
	Contains(Node) bool
	Len() int
}

type nodeSet map[Node]struct{}

func (ns nodeSet) Iterate(f func(Node) bool) bool {
	for node := range ns {
		if !f(node) {
			return false
		}
	}
	return true
}

func (ns nodeSet) Each(f func(Node)) {
	_nodesEach(ns, f)
}

func (ns nodeSet) Filter(f func(Node) bool) NodeCollection {
	return _nodesFilter(ns, f)
}

func (ns nodeSet) ToArray() NodeSlice {
	return _nodesToArray(ns)
}

func (ns nodeSet) ToSet() NodeSet {
	return ns
}

func (ns nodeSet) Contains(node Node) bool {
	_, ok := ns[node]
	return ok
}

func (ns nodeSet) Len() int {
	return len(ns)
}

func (ns nodeSet) Add(node Node) {
	if node == nil {
		return
	}

	ns[node] = struct{}{}
}

func (ns nodeSet) Remove(node Node) {
	if node == nil {
		return
	}

	delete(ns, node)
}
