package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stronglyConnectedComponents(t *testing.T) {

	tests := []struct {
		name  string
		edges [][2]Node
		scc   [][]Node
	}{
		{
			"case 1",
			[][2]Node{
				{1, 2}, {2, 3}, {2, 8}, {3, 4}, {3, 7}, {4, 5}, {5, 3}, {5, 6}, {7, 4}, {7, 6}, {8, 1}, {8, 7},
			},
			[][]Node{
				{3, 4, 5, 7}, {1, 2, 8}, {6},
			},
		},
		{
			"case 2",
			[][2]Node{
				{1, 2}, {1, 3}, {1, 4}, {4, 2}, {3, 4}, {2, 3},
			},
			[][]Node{
				{2, 3, 4}, {1},
			},
		},
		{
			"case 3",
			[][2]Node{
				{1, 2}, {2, 3}, {3, 2}, {2, 1},
			},
			[][]Node{
				{1, 2, 3},
			},
		},
		{
			"case 4",
			[][2]Node{
				{0, 1}, {1, 2}, {1, 3}, {2, 4}, {2, 5}, {3, 4}, {3, 5}, {4, 6},
			},
			[][]Node{
				{0}, {1}, {2}, {3}, {4}, {5}, {6},
			},
		},
		{
			"case 5",
			[][2]Node{
				{0, 1}, {1, 2}, {1, 3}, {1, 4}, {2, 0}, {2, 3}, {3, 4}, {4, 3},
			},
			[][]Node{
				{0, 1, 2}, {3, 4},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDirectedGraph()
			for _, e := range tt.edges {
				AddEdge(g, e[0], e[1])
			}

			toSet := func(scc [][]Node) []map[Node]struct{} {
				var res []map[Node]struct{}
				for _, i := range scc {
					m := make(map[Node]struct{})
					for _, j := range i {
						m[j] = struct{}{}
					}
					res = append(res, m)
				}
				return res
			}

			scc := stronglyConnectedComponents(g)
			//ctx := sccContext{g: g}
			//scc := ctx.FindScc()
			s1 := toSet(tt.scc)
			s2 := toSet(scc)
			assert.Subset(t, s1, s2)
			assert.Subset(t, s2, s1)
		})
	}
}

func isCyclicPermutation(a, b []Node) bool {
	l := len(a)
	if len(b) != l {
		return false
	}
	if l == 0 {
		return true
	}

	var aa []Node
	aa = append(aa, a...)
	aa = append(aa, a...)

	for i := 0; i < l; i += 1 {
		eq := true
		for j := 0; j < l; j += 1 {
			if aa[i+j] != b[j] {
				eq = false
				break
			}
		}
		if eq {
			return true
		}
	}
	return false
}

func Test_isCyclicPermutation(t *testing.T) {
	tests := []struct {
		name string
		a    []Node
		b    []Node
		want bool
	}{
		{"case 1", []Node{}, []Node{}, true},
		{"case 2", []Node{1}, []Node{1}, true},
		{"case 3", []Node{1}, []Node{1, 2}, false},
		{"case 4", []Node{1, 2}, []Node{2, 1}, true},
		{"case 5", []Node{1, 2, 3, 4, 5}, []Node{3, 4, 5, 1, 2}, true},
		{"case 5", []Node{1, 2, 3, 4, 5}, []Node{3, 4, 5, 2, 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := isCyclicPermutation(tt.a, tt.b)
			assert.Equal(t, tt.want, res)
		})
	}
}

func assertCyclesEqual(t *testing.T, a, b []Cycle) {
	if len(a) != len(b) {
		t.Errorf("%v is not equal to %v", a, b)
		return
	}

	for _, cycle1 := range a {
		found := false
		for _, cycle2 := range b {
			if isCyclicPermutation(cycle1, cycle2) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%v is not in %v", cycle1, b)
			return
		}
	}
}

func Test_assertCyclesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []Cycle
		b    []Cycle
	}{
		{"case1", []Cycle{}, []Cycle{}},
		{"case2", []Cycle{{}, {}}, []Cycle{{}, {}}},
		{"case2", []Cycle{{1, 2, 3}, {4, 5}}, []Cycle{{5, 4}, {3, 1, 2}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCyclesEqual(t, tt.a, tt.b)
		})
	}
}

func addCycle(g DirectedGraph, cycle []Node) {
	if len(cycle) == 0 {
		return
	}

	for i := 0; i < len(cycle)-1; i += 1 {
		AddEdge(g, cycle[i], cycle[i+1])
	}
	AddEdge(g, cycle[len(cycle)-1], cycle[0])
}

func worstCaseGraph(k int) DirectedGraph {
	g := NewDirectedGraph()
	for n := 2; n < k+2; n += 1 {
		AddEdge(g, 1, n)
		AddEdge(g, n, k+2)
	}
	AddEdge(g, 2*k+1, 1)

	for n := k + 2; n < 2*k+2; n += 1 {
		AddEdge(g, n, 2*k+2)
		AddEdge(g, n, n+1)
	}
	AddEdge(g, 2*k+3, k+2)

	for n := 2*k + 3; n < 3*k+3; n += 1 {
		AddEdge(g, 2*k+2, n)
		AddEdge(g, n, 3*k+3)
	}
	AddEdge(g, 3*k+3, 2*k+2)
	return g
}

func TestFindCycles(t *testing.T) {
	t.Run("case 1", func(t *testing.T) {
		edges := [][2]Node{{0, 0}, {0, 1}, {0, 2}, {1, 2}, {2, 0}, {2, 1}, {2, 2}}
		g := NewDirectedGraph()
		AddEdges(g, edges)

		cycles := FindCycles(g)
		expectedCycles := []Cycle{{0}, {0, 1, 2}, {0, 2}, {1, 2}, {2}}
		assertCyclesEqual(t, expectedCycles, cycles)
	})

	t.Run("case 2", func(t *testing.T) {
		g := NewDirectedGraph()
		addCycle(g, []Node{1, 2, 3})
		expectedCycles := []Cycle{{1, 2, 3}}
		assertCyclesEqual(t, expectedCycles, FindCycles(g))

		addCycle(g, []Node{10, 20, 30})
		expectedCycles2 := []Cycle{{1, 2, 3}, {10, 20, 30}}
		assertCyclesEqual(t, expectedCycles2, FindCycles(g))
	})

	t.Run("case 3", func(t *testing.T) {
		for k := 3; k < 10; k += 1 {
			g := worstCaseGraph(k)
			l := len(FindCycles(g))
			assert.Equal(t, 3*k, l)
		}
	})
}
