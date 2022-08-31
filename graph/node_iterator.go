package graph

import "github.com/jison/uni/internal/reflecting"

type Node interface{}

func canBeNode(v Node) bool {
	return v != nil && reflecting.CanBeMapKey(v)
}

type NodeAndAttrsIterator interface {
	Iterate(func(node Node, attrs AttrsView) bool) bool
}

type NodeIterator interface {
	Iterate(func(Node) bool) bool
}

type nodeIterateFunc func(f func(node Node) bool) bool

func (i nodeIterateFunc) Iterate(f func(Node) bool) bool {
	if i == nil {
		return true
	}
	return i(f)
}

type NodesWithAttrs map[Node]Attrs

func (na NodesWithAttrs) Iterate(f func(node Node, attrs AttrsView) bool) bool {
	for n, v := range na {
		if !f(n, v) {
			return false
		}
	}
	return true
}

func NodesWithAttrsFrom(ni NodeAndAttrsIterator) NodesWithAttrs {
	nodes := NodesWithAttrs{}
	if ni == nil {
		return nodes
	}
	ni.Iterate(func(node Node, attrs AttrsView) bool {
		var nodeAttrs Attrs
		var ok bool
		if nodeAttrs, ok = nodes[node]; !ok {
			nodeAttrs = Attrs{}
			nodes[node] = nodeAttrs
		}

		attrs.Iterate(func(key interface{}, value interface{}) bool {
			nodeAttrs.Set(key, value)
			return true
		})

		return true
	})

	return nodes
}

type graphNodes struct {
	graph DirectedGraphView
	nodes NodeIterator
}

func (gn *graphNodes) Iterate(f func(node Node, attrs AttrsView) bool) bool {
	if gn == nil || gn.graph == nil {
		return true
	}

	if gn.nodes == nil {
		return true
	}

	return gn.nodes.Iterate(func(node Node) bool {
		attrs, ok := gn.graph.NodeAttrs(node)
		if !ok {
			return true
		}
		return f(node, attrs)
	})
}

type emptyNodeIterator struct{}

func (e emptyNodeIterator) Iterate(_ func(node Node, attrs AttrsView) bool) bool { return true }

type NodeEntry struct {
	node  Node
	attrs AttrsView
}

func (ne *NodeEntry) Iterate(f func(node Node, attrs AttrsView) bool) bool {
	if ne == nil {
		return true
	}
	return f(ne.node, ne.attrs)
}

type combinedNodeIterator struct {
	left  NodeAndAttrsIterator
	right NodeAndAttrsIterator
}

func (i *combinedNodeIterator) Iterate(f func(node Node, attrs AttrsView) bool) bool {
	if i == nil {
		return true
	}
	if i.left != nil {
		if !i.left.Iterate(f) {
			return false
		}
	}
	if i.right != nil {
		return i.right.Iterate(f)
	}

	return true
}

func combineNodeIterators(iterators ...NodeAndAttrsIterator) NodeAndAttrsIterator {
	if len(iterators) == 0 {
		return emptyNodeIterator{}
	}

	combined := iterators[0]
	for _, i := range iterators[1:] {
		if i == nil {
			continue
		}
		if _, ok := i.(emptyNodeIterator); ok {
			continue
		}
		combined = &combinedNodeIterator{combined, i}
	}

	return combined
}
