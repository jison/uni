package graph

type EdgeEntry struct {
	from  Node
	to    Node
	attrs AttrsView
}

type EdgeAndAttrsIterator interface {
	Iterate(func(from Node, to Node, attrs AttrsView) bool) bool
}

type EdgesWithAttrs map[Node]map[Node]Attrs

func (ei EdgesWithAttrs) Iterate(f func(from Node, to Node, attrs AttrsView) bool) bool {
	for from, na := range ei {
		for to, attrs := range na {
			if !f(from, to, attrs) {
				return false
			}
		}
	}
	return true
}

type edgeIterator struct {
	edges            EdgesWithAttrs
	reverseDirection bool
	restrictedNodes  nodeIterateFunc
}

func (ei *edgeIterator) Iterate(f func(from Node, to Node, attrs AttrsView) bool) bool {
	ff := f
	if ei.reverseDirection {
		ff = func(from Node, to Node, attrs AttrsView) bool {
			return f(to, from, attrs)
		}
	}

	if ei.restrictedNodes == nil {
		for from, na := range ei.edges {
			for to, attrs := range na {
				if !ff(from, to, attrs) {
					return false
				}
			}
		}
		return true
	} else {
		return ei.restrictedNodes.Iterate(func(from Node) bool {
			for to, attrs := range ei.edges[from] {
				if !ff(from, to, attrs) {
					return false
				}
			}
			return true
		})
	}
}

func EdgesWithAttrsFrom(ei EdgeAndAttrsIterator) EdgesWithAttrs {
	edges := EdgesWithAttrs{}
	if ei == nil {
		return edges
	}
	ei.Iterate(func(from Node, to Node, attrs AttrsView) bool {
		var toNodes map[Node]Attrs
		var ok bool
		if toNodes, ok = edges[from]; !ok {
			toNodes = map[Node]Attrs{}
			edges[from] = toNodes
		}

		var edgeAttrs Attrs
		if edgeAttrs, ok = toNodes[to]; !ok {
			edgeAttrs = Attrs{}
			toNodes[to] = edgeAttrs
		}

		attrs.Iterate(func(key interface{}, value interface{}) bool {
			edgeAttrs.Set(key, value)
			return true
		})
		return true
	})

	return edges
}

type emptyEdgeIterator struct{}

func (e emptyEdgeIterator) Iterate(_ func(from Node, to Node, attrs AttrsView) bool) bool {
	return true
}
