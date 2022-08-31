package graph

func NewDirectedGraphWithInNodes(source DirectedGraphView, nodes []Node) DirectedGraph {
	subG := SubGraphWithNodes(source, nodes)
	return NewDirectedGraphFrom(subG)
}

func GetNodesInDirectionMatch(node Node, directionFunc func(Node) NodeAndAttrsIterator,
	predicate func(Node, AttrsView) bool) NodeAndAttrsIterator {

	it := getNodesInDirectionMatch(node, directionFunc, predicate, map[Node]struct{}{})
	return NodesWithAttrsFrom(it) // remove duplicate nodes
}

func getNodesInDirectionMatch(node Node, directionFunc func(Node) NodeAndAttrsIterator,
	predicate func(Node, AttrsView) bool, visitedNodes map[Node]struct{}) NodeAndAttrsIterator {

	var it NodeAndAttrsIterator = emptyNodeIterator{}

	directionFunc(node).Iterate(func(subNode Node, attrs AttrsView) bool {
		if _, visited := visitedNodes[subNode]; visited {
			return true
		}

		if predicate(subNode, attrs) {
			it = combineNodeIterators(it, &NodeEntry{
				node:  subNode,
				attrs: attrs,
			})
		} else {
			visitedNodes[subNode] = struct{}{}
			subIt := getNodesInDirectionMatch(subNode, directionFunc, predicate, visitedNodes)
			it = combineNodeIterators(it, subIt)
		}

		return true
	})

	return it
}
