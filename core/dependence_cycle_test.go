package core

import (
	"fmt"
	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/graph"
	"strings"
	"testing"

	"github.com/jison/uni/core/model"
	"github.com/stretchr/testify/assert"
)

func graphWithoutCycle() DependenceGraph {
	m := model.NewModule(
		model.Func(func(int) string {
			return ""
		}),
		model.Func(func(int) byte {
			return 0
		}),
		model.Func(func(byte) rune {
			return 0
		}),
	)
	rep := model.NewRepository(m.AllComponents())
	return newDependenceGraph(rep)
}

func graphWithOneCycle() (DependenceGraph, model.ComponentRepository) {
	m := model.NewModule(
		model.Func(func(int) string {
			return ""
		}, model.Return(0, model.Name("name1"))),
		model.Func(func(string) int {
			return 0
		}, model.Return(0, model.Name("name2"))),
	)
	rep := model.NewRepository(m.AllComponents())
	return newDependenceGraph(rep), rep
}

func graphWithCrossingCycles() (DependenceGraph, model.ComponentRepository) {
	m := model.NewModule(
		model.Func(func(int) string {
			return ""
		}, model.Return(0, model.Name("name1"))),
		model.Func(func(string) int {
			return 0
		}, model.Return(0, model.Name("name2"))),
		model.Func(func(int) byte {
			return 0
		}, model.Return(0, model.Name("name3"))),
		model.Func(func(byte) string {
			return ""
		}, model.Return(0, model.Name("name4"))),
		model.Func(func(string) byte {
			return 0
		}, model.Return(0, model.Name("name5"))),
	)
	rep := model.NewRepository(m.AllComponents())
	return newDependenceGraph(rep), rep
}

func pairsOfCycle(cycle DependenceCycle) map[[2]Node]struct{} {
	nodes := NewNodeCollection(cycle.Nodes()).ToArray()

	pairs := map[[2]Node]struct{}{}
	for i, n := range nodes {
		var pair [2]Node
		if i < len(nodes)-1 {
			pair = [2]Node{n, nodes[i+1]}
		} else {
			pair = [2]Node{n, nodes[0]}
		}
		pairs[pair] = struct{}{}
	}

	return pairs
}

func verifyCycleIsInGraph(t *testing.T, g DependenceGraph, cycle DependenceCycle) {
	pairs := pairsOfCycle(cycle)

	nodes := NewNodeCollection(cycle.Nodes()).ToArray()
	assert.Equal(t, len(nodes), len(pairs))

	for p := range pairs {
		assert.True(t, NewNodeCollection(g.InputNodesTo(p[1])).ToSet().Contains(p[0]))
	}
}

func cycleEqual(cycle1 DependenceCycle, cycle2 DependenceCycle) bool {
	p1 := pairsOfCycle(cycle1)
	p2 := pairsOfCycle(cycle2)

	return assert.ObjectsAreEqual(p1, p2)
}

func verifyCyclesOfNode(t *testing.T, g DependenceGraph, node Node, count int) {
	cycles := g.CycleInfo().CyclesOfNode(node)
	assert.Equal(t, count, len(cycles))
	for _, cycle := range cycles {
		firstNode := NewNodeCollection(cycle.Nodes()).ToArray()[0]
		assert.Same(t, node, firstNode)
		verifyCycleIsInGraph(t, g, cycle)
	}
}

func verifyComponentsInOneCycle(t *testing.T, g DependenceGraph, rep model.ComponentRepository, names ...string) {
	verifyIsInputCom := func(com model.Component, inCom model.Component) {
		inputComs := g.InputComponentsTo(com).ToSet()
		assert.True(t, inputComs.Contains(inCom))
	}

	firstCom := comByName(rep, names[0])
	lastCom := firstCom
	for _, name := range names[1:] {
		curCom := comByName(rep, name)
		verifyIsInputCom(lastCom, curCom)
		lastCom = curCom
	}
	verifyIsInputCom(lastCom, firstCom)
}

func Test_buildCycleInfoOf(t *testing.T) {
	t.Run("no cycles", func(t *testing.T) {
		g := graphWithoutCycle()
		cycleInfo := buildCycleInfoOf(g)
		assert.Equal(t, 0, len(cycleInfo.Cycles()))

		g.Nodes().Iterate(func(node Node) bool {
			assert.Equal(t, 0, len(cycleInfo.CyclesOfNode(node)))
			return true
		})
	})

	t.Run("one cycle", func(t *testing.T) {
		g, _ := graphWithOneCycle()
		cycleInfo := buildCycleInfoOf(g)
		assert.Equal(t, 1, len(cycleInfo.Cycles()))

		cycle := cycleInfo.Cycles()[0]
		s1 := NewNodeCollection(cycle.Nodes()).ToSet()
		s2 := g.Nodes().ToSet()
		assert.Equal(t, s2, s1)
		verifyCycleIsInGraph(t, g, cycle)

		g.Nodes().Iterate(func(node Node) bool {
			verifyCyclesOfNode(t, g, node, 1)
			return true
		})
	})

	t.Run("cycles with crossing node", func(t *testing.T) {
		g, rep := graphWithCrossingCycles()

		cycleInfo := buildCycleInfoOf(g)
		assert.Equal(t, 4, len(cycleInfo.Cycles()))
		for _, cycle := range cycleInfo.Cycles() {
			verifyCycleIsInGraph(t, g, cycle)
		}

		verifyCyclesOfNode(t, g, comByName(rep, "name1").Valuer(), 2)
		verifyCyclesOfNode(t, g, comByName(rep, "name2").Valuer(), 3)
		verifyCyclesOfNode(t, g, comByName(rep, "name3").Valuer(), 1)
		verifyCyclesOfNode(t, g, comByName(rep, "name4").Valuer(), 3)
		verifyCyclesOfNode(t, g, comByName(rep, "name5").Valuer(), 2)

		verifyComponentsInOneCycle(t, g, rep, "name1", "name2")
		verifyComponentsInOneCycle(t, g, rep, "name2", "name4", "name3")
		verifyComponentsInOneCycle(t, g, rep, "name4", "name5")
		verifyComponentsInOneCycle(t, g, rep, "name2", "name4", "name5", "name1")
	})
}

func Test_dependenceCycleNode(t *testing.T) {
	g, rep := graphWithOneCycle()

	cycleInfo := buildCycleInfoOf(g)
	com1 := comByName(rep, "name1")
	cycle := cycleInfo.CyclesOfNode(com1.Valuer())[0]

	var nodes []Node
	nodes = append(nodes, com1.Valuer())

	node := g.InputNodesTo(com1.Valuer()).ToArray()[0]
	for node != com1.Valuer() {
		nodes = append(nodes, node)
		node = g.InputNodesTo(node).ToArray()[0]
	}

	t.Run("Iterate", func(t *testing.T) {
		testNodeIterator(t, cycle.Nodes(), nodes)
	})

	t.Run("Format", func(t *testing.T) {
		var dep1 model.Dependency
		com1.Provider().Dependencies().Iterate(func(d model.Dependency) bool {
			dep1 = d
			return false
		})
		com2 := comByName(rep, "name2")
		var dep2 model.Dependency
		com2.Provider().Dependencies().Iterate(func(d model.Dependency) bool {
			dep2 = d
			return false
		})

		t.Run("not verbose", func(t *testing.T) {
			expected := strings.Builder{}
			expected.WriteString("cycle:")
			expected.WriteString(fmt.Sprintf("\n\t%v", com1))
			expected.WriteString(fmt.Sprintf("\n\t%v", dep2))
			expected.WriteString(fmt.Sprintf("\n\t%+v", com2.Provider()))
			expected.WriteString(fmt.Sprintf("\n\t%v", com2))
			expected.WriteString(fmt.Sprintf("\n\t%v", dep1))
			expected.WriteString(fmt.Sprintf("\n\t%+v", com1.Provider()))

			str := fmt.Sprintf("%v", cycle)
			assert.Equal(t, expected.String(), str)
		})

		t.Run("verbose", func(t *testing.T) {
			expected := strings.Builder{}
			expected.WriteString("cycle:")
			expected.WriteString(fmt.Sprintf("\n\t%v", com1))
			expected.WriteString(fmt.Sprintf("\n\t%v", dep2))
			expected.WriteString(fmt.Sprintf("\n\t%+v", com2.Provider()))
			expected.WriteString(fmt.Sprintf("\n\t%v", com2))
			expected.WriteString(fmt.Sprintf("\n\t%v", dep1))
			expected.WriteString(fmt.Sprintf("\n\t%+v", com1.Provider()))

			str := fmt.Sprintf("%+v", cycle)
			assert.Equal(t, expected.String(), str)
		})
	})
}

func Test_dependenceCycleFromGCycle(t *testing.T) {
	dg := newDependenceGraphForTest(model.EmptyComponents())

	t.Run("cycle is empty", func(t *testing.T) {
		r := dependenceCycleFromGCycle(dg, []graph.Node{})
		assert.Nil(t, r)
	})

	t.Run("node in cycle in are not valuer", func(t *testing.T) {
		r := dependenceCycleFromGCycle(dg, []graph.Node{1, 2, 3})
		assert.Nil(t, r)
	})

	t.Run("normal cycle", func(t *testing.T) {
		node1 := valuer.Identity()
		node2 := valuer.Identity()

		nodes := []graph.Node{node1, 123, node2}
		r := dependenceCycleFromGCycle(dg, nodes)
		pairs := pairsOfCycle(r)

		pairs2 := map[[2]Node]struct{}{
			{node1, node2}: {},
			{node2, node1}: {},
		}
		assert.Equal(t, pairs, pairs2)
	})
}
