package core

import "fmt"

type Path interface {
	Graph() DependenceGraph
	Nodes() NodeIterator
	Len() int
	Contains(node Node) bool
	Append(node Node) Path
	Reversed() Path
}

func NewPath(g DependenceGraph) Path {
	return &pathNode{
		graph: g,
	}
}

type pathNode struct {
	graph   DependenceGraph
	prev    *pathNode
	node    Node
	nodeNum int
}

func (p *pathNode) Graph() DependenceGraph {
	if p == nil {
		return nil
	} else {
		return p.graph
	}
}

func (p *pathNode) Nodes() NodeIterator {
	return p
}

func (p *pathNode) Len() int {
	if p == nil {
		return 0
	}

	return p.nodeNum
}

func (p *pathNode) Contains(node Node) bool {
	if p == nil || node == nil {
		return false
	}

	return !p.Iterate(func(n Node) bool {
		return n != node
	})
}

func (p *pathNode) Append(node Node) Path {
	if node == nil {
		return p
	}

	if p == nil {
		return &pathNode{
			graph:   nil,
			prev:    nil,
			node:    node,
			nodeNum: 1,
		}
	} else {
		return &pathNode{
			graph:   p.graph,
			prev:    p,
			node:    node,
			nodeNum: p.nodeNum + 1,
		}
	}
}

func (p *pathNode) Reversed() Path {
	var prev Path
	if p == nil {
		prev = NewPath(nil)
	} else {
		prev = NewPath(p.graph)
	}

	p.Iterate(func(node Node) bool {
		prev = prev.Append(node)
		return true
	})

	return prev
}

func (p *pathNode) Iterate(f func(node Node) bool) bool {
	first := p
	cur := p
	for cur != nil {
		if cur.node != nil && cur.nodeNum > 0 {
			if !f(cur.node) {
				return false
			}
		}

		cur = cur.prev
		if cur == first {
			break
		}
	}
	return true
}

func (p *pathNode) Format(fs fmt.State, r rune) {
	if p.Len() == 0 {
		_, _ = fmt.Fprint(fs, "empty path")
	} else {
		_, _ = fmt.Fprint(fs, "path:\n")
		_formatNodes(p.graph, p, fs, r)
	}
}

func _formatNodes(graph DependenceGraph, ni NodeIterator, fs fmt.State, r rune) {
	verbose := fs.Flag('+') && r == 'v'

	ni.Iterate(func(node Node) bool {
		if dep, ok := graph.DependencyOfNode(node); ok {
			_, _ = fmt.Fprintf(fs, "\t%v\n", dep)
			return true
		}

		if com, ok := graph.ComponentOfNode(node); ok {
			_, _ = fmt.Fprintf(fs, "\t%v\n", com)
			return true
		}

		if pro, ok := graph.ProviderOfNode(node); ok {
			_, _ = fmt.Fprintf(fs, "\t%+v\n", pro)
			return true
		}

		if con, ok := graph.ConsumerOfNode(node); ok {
			_, _ = fmt.Fprintf(fs, "\t%+v\n", con)
			return true
		}

		if !verbose {
			return true
		}

		_, _ = fmt.Fprintf(fs, "\t%v\n", node)

		return true
	})
}
