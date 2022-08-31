package graph

// stronglyConnectedComponents
// find strongly connected components of graph
// see https://github.com/networkx/networkx/blob/main/networkx/algorithms/components/strongly_connected.py#L16
func stronglyConnectedComponents(g DirectedGraph) [][]Node {
	min := func(a, b int) int {
		if a <= b {
			return a
		} else {
			return b
		}
	}
	mapContains := func(m map[Node]int, n Node) bool {
		if _, ok := m[n]; ok {
			return true
		} else {
			return false
		}
	}

	preorder := make(map[Node]int)
	lowLink := make(map[Node]int)
	sccFound := make(map[Node]struct{})
	var sccQueue []Node

	i := 0

	var allScc [][]Node

	g.Nodes().Iterate(func(source Node, _ AttrsView) bool {
		if _, ok := sccFound[source]; ok {
			return true
		}

		queue := []Node{source}
		for len(queue) > 0 {
			node := queue[len(queue)-1]
			if !mapContains(preorder, node) {
				i += 1
				preorder[node] = i
			}
			done := true
			SuccessorsOf(g, node).Iterate(func(suc Node, _ AttrsView) bool {
				if !mapContains(preorder, suc) {
					queue = append(queue, suc)
					done = false
					return false
				}
				return true
			})

			if done {
				lowLink[node] = preorder[node]
				SuccessorsOf(g, node).Iterate(func(suc Node, _ AttrsView) bool {
					if _, ok := sccFound[suc]; ok {
						return true
					}
					if preorder[suc] > preorder[node] {
						lowLink[node] = min(lowLink[node], lowLink[suc])
					} else {
						lowLink[node] = min(lowLink[node], preorder[suc])
					}
					return true
				})

				if lowLink[node] == preorder[node] {
					scc := []Node{node}

					for len(sccQueue) > 0 && preorder[sccQueue[len(sccQueue)-1]] > preorder[node] {
						k := sccQueue[len(sccQueue)-1]
						sccQueue = sccQueue[0 : len(sccQueue)-1]
						scc = append(scc, k)
					}
					for _, sccV := range scc {
						sccFound[sccV] = struct{}{}
					}

					allScc = append(allScc, scc)
				} else {
					sccQueue = append(sccQueue, node)
				}

				queue = queue[0 : len(queue)-1]
			}
		}
		return true
	})

	return allScc
}

type Cycle []Node

type NodeSet map[Node]struct{}

func (s NodeSet) Add(node Node) {
	s[node] = struct{}{}
}

func (s NodeSet) Has(node Node) bool {
	_, ok := s[node]
	return ok
}

func (s NodeSet) Del(node Node) {
	delete(s, node)
}

func (s NodeSet) Len() int {
	return len(s)
}

// FindCycles
// see https://github.com/networkx/networkx/blob/main/networkx/algorithms/cycles.py#L98
func FindCycles(g DirectedGraphView) []Cycle {
	_unblock := func(node Node, blocked NodeSet, B map[Node]NodeSet) {
		stack := []Node{node}
		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[0 : len(stack)-1]
			if blocked.Has(n) {
				blocked.Del(n)
				for nn := range B[n] {
					stack = append(stack, nn)
				}
				B[n] = NodeSet{}
			}
		}
	}

	type stackItem struct {
		node      Node
		neighbors []Node
	}

	_pushStack := func(stack []*stackItem, g DirectedGraph, node Node) []*stackItem {
		var neighbors []Node
		SuccessorsOf(g, node).Iterate(func(n Node, _ AttrsView) bool {
			neighbors = append(neighbors, n)
			return true
		})
		return append(stack, &stackItem{
			node:      node,
			neighbors: neighbors,
		})
	}

	subG := NewDirectedGraph()
	g.Edges().Iterate(func(from Node, to Node, attrs AttrsView) bool {
		AddEdge(subG, from, to)
		return true
	})
	allScc := stronglyConnectedComponents(subG)
	//goland:noinspection SpellCheckingInspection
	var sccs [][]Node
	for _, scc := range allScc {
		if len(scc) > 1 {
			sccs = append(sccs, scc)
		}
	}

	var cycles []Cycle
	subG.Nodes().Iterate(func(node Node, _ AttrsView) bool {
		if HasEdge(subG, node, node) {
			cycles = append(cycles, []Node{node})
			subG.RemoveEdge(node, node)
		}
		return true
	})

	for len(sccs) > 0 {
		scc := sccs[len(sccs)-1]
		sccs = sccs[0 : len(sccs)-1]

		sccG := NewDirectedGraphWithInNodes(subG, scc)

		startNode := scc[len(scc)-1]
		scc = scc[0 : len(scc)-1]

		path := []Node{startNode}
		blocked := NodeSet{}
		closed := NodeSet{}
		blocked.Add(startNode)
		B := make(map[Node]NodeSet)
		stack := _pushStack([]*stackItem{}, sccG, startNode)
		for len(stack) > 0 {
			item := stack[len(stack)-1]

			if len(item.neighbors) > 0 {
				nextNode := item.neighbors[len(item.neighbors)-1]
				item.neighbors = item.neighbors[0 : len(item.neighbors)-1]
				if nextNode == startNode {
					var cycle []Node
					cycle = append(cycle, path...)
					cycles = append(cycles, cycle)
					for _, node := range cycle {
						closed.Add(node)
					}
				} else if !blocked.Has(nextNode) {
					path = append(path, nextNode)
					stack = _pushStack(stack, sccG, nextNode)
					closed.Del(nextNode)
					blocked.Add(nextNode)
					continue
				}
			}

			if len(item.neighbors) == 0 {
				thisNode := item.node
				if closed.Has(thisNode) {
					_unblock(thisNode, blocked, B)
				} else {
					SuccessorsOf(sccG, thisNode).Iterate(func(nbr Node, _ AttrsView) bool {
						set, ok := B[nbr]
						if !ok {
							set = NodeSet{}
							B[nbr] = set
						}
						if !set.Has(thisNode) {
							set.Add(thisNode)
						}
						return true
					})
				}
				stack = stack[0 : len(stack)-1]
				path = path[0 : len(path)-1]
			}
		}

		H := NewDirectedGraphWithInNodes(subG, scc)
		for _, subScc := range stronglyConnectedComponents(H) {
			if len(subScc) > 1 {
				sccs = append(sccs, subScc)
			}
		}
	}

	return cycles
}
