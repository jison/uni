package core

import (
	"fmt"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/graph"
)

type DependenceCycle interface {
	Nodes() NodeIterator
}

type DependenceCycleInfo interface {
	CyclesOfNode(node Node) []DependenceCycle
	Cycles() []DependenceCycle
}

type dependenceCycleInfo struct {
	cycles       []DependenceCycle
	cyclesByNode map[Node][]DependenceCycle
}

func (ci *dependenceCycleInfo) CyclesOfNode(node Node) []DependenceCycle {
	return ci.cyclesByNode[node]
}

func (ci *dependenceCycleInfo) Cycles() []DependenceCycle {
	return ci.cycles
}

func buildCycleInfoOf(dg DependenceGraph) DependenceCycleInfo {
	gCycles := graph.FindCycles(dg.Graph())

	var cycles []DependenceCycle
	cyclesByNode := map[Node][]DependenceCycle{}
	for _, gCycle := range gCycles {
		if len(gCycle) == 0 {
			continue
		}

		var headCycleNode *dependenceCycleNode
		var tailCycleNode *dependenceCycleNode
		for _, gNode := range gCycle {
			valNode, ok := gNode.(valuer.Valuer)
			if !ok {
				continue
			}

			currentNode := &dependenceCycleNode{
				graph: dg,
				node:  valNode,
				next:  nil,
			}

			cyclesByNode[valNode] = append(cyclesByNode[valNode], currentNode)

			if headCycleNode == nil {
				headCycleNode = currentNode
			}
			if tailCycleNode != nil {
				tailCycleNode.next = currentNode
			}
			tailCycleNode = currentNode
		}
		tailCycleNode.next = headCycleNode

		cycles = append(cycles, headCycleNode)
	}

	return &dependenceCycleInfo{
		cycles:       cycles,
		cyclesByNode: cyclesByNode,
	}
}

type dependenceCycleNode struct {
	graph DependenceGraph
	node  Node
	next  *dependenceCycleNode
}

func (c *dependenceCycleNode) Nodes() NodeIterator {
	return c
}

func (c *dependenceCycleNode) Iterate(f func(Node) bool) bool {
	if !f(c.node) {
		return false
	}

	cur := c.next
	for cur != c {
		if !f(cur.node) {
			return false
		}
		cur = cur.next
	}

	return true
}

func (c *dependenceCycleNode) Format(fs fmt.State, r rune) {
	_, _ = fmt.Fprint(fs, "cycle:\n")
	_formatNodes(c.graph, c, fs, r)
}
