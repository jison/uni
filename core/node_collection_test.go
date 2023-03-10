package core

import (
	"math/rand"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/stretchr/testify/assert"
)

func testNodeIterator(t *testing.T, ni NodeIterator, nodes []Node) {
	t.Run("iterate", func(t *testing.T) {
		m1 := map[Node]struct{}{}
		m2 := map[Node]struct{}{}

		for _, n := range nodes {
			m1[n] = struct{}{}
		}

		r := ni.Iterate(func(node Node) bool {
			m2[node] = struct{}{}
			return true
		})

		assert.True(t, r)
		assert.Equal(t, m1, m2)
	})

	t.Run("interrupt", func(t *testing.T) {
		if len(nodes) == 0 {
			n := 0
			r := ni.Iterate(func(node Node) bool {
				n += 1
				return false
			})
			assert.True(t, r)
			assert.Equal(t, 0, n)
		} else {
			var halfNodes []Node
			r := ni.Iterate(func(node Node) bool {
				halfNodes = append(halfNodes, node)
				return len(halfNodes) < len(nodes)/2
			})

			assert.False(t, r)

			expected := len(nodes) / 2
			if expected == 0 {
				expected = 1
			}
			assert.Equal(t, expected, len(halfNodes))
		}
	})
}

func Test_filteredNodeIterator(t *testing.T) {
	nodes := []Node{valuer.Identity(), valuer.Identity(), valuer.Identity(), valuer.Identity()}
	set := map[Node]struct{}{}
	var filteredNodes []Node
	for _, n := range nodes {
		if rand.Intn(2) == 0 {
			continue
		}
		set[n] = struct{}{}
		filteredNodes = append(filteredNodes, n)
	}

	predicate := func(node Node) bool {
		_, ok := set[node]
		return ok
	}

	type args struct {
		original  NodeIterator
		predicate func(Node) bool
	}

	tests := []struct {
		name  string
		args  args
		nodes []Node
	}{
		{"original is nil", args{nil, predicate}, []Node{}},
		{"predicate is nil", args{NodeSlice(nodes), nil}, nodes},
		{"predicate is nil", args{NodeSlice(nodes), predicate}, filteredNodes},
	}

	for _, tt := range tests {
		it := &filteredNodeIterator{tt.args.original, tt.args.predicate}
		testNodeIterator(t, it, tt.nodes)
	}
}

func testNodeCollection(t *testing.T, nc NodeCollection, nodes []Node, isRecursive bool) {
	testNodeIterator(t, nc, nodes)

	t.Run("Each", func(t *testing.T) {
		m1 := map[Node]struct{}{}
		m2 := map[Node]struct{}{}

		for _, n := range nodes {
			m1[n] = struct{}{}
		}

		nc.Each(func(node Node) {
			m2[node] = struct{}{}
		})

		assert.Equal(t, m1, m2)
	})

	t.Run("Filter", func(t *testing.T) {
		if isRecursive {
			return
		}
		set := map[Node]struct{}{}
		var filteredNodes []Node
		for _, n := range nodes {
			if rand.Intn(2) == 0 {
				continue
			}
			set[n] = struct{}{}
			filteredNodes = append(filteredNodes, n)
		}

		filteredNc := nc.Filter(func(node Node) bool {
			_, ok := set[node]
			return ok
		})
		testNodeCollection(t, filteredNc, filteredNodes, true)
	})

	t.Run("ToArray", func(t *testing.T) {
		m1 := map[Node]struct{}{}
		m2 := map[Node]struct{}{}

		for _, n := range nodes {
			m1[n] = struct{}{}
		}
		for _, n := range nc.ToArray() {
			m2[n] = struct{}{}
		}

		assert.Equal(t, m1, m2)
	})

	t.Run("ToSet", func(t *testing.T) {
		if isRecursive {
			return
		}
		testNodeSet(t, nc.ToSet(), nodes, true)
	})
}

func testNodeSet(t *testing.T, ns NodeSet, nodes []Node, isRecursive bool) {
	testNodeCollection(t, ns, nodes, isRecursive)

	t.Run("Contains", func(t *testing.T) {
		for _, n := range nodes {
			assert.True(t, ns.Contains(n))
		}

		for i := 0; i < 10; i++ {
			c := valuer.Identity()
			assert.False(t, ns.Contains(c))
		}
	})

	t.Run("Len", func(t *testing.T) {
		assert.Equal(t, len(nodes), ns.Len())
	})
}

func Test_nodeCollection(t *testing.T) {
	nodes := []Node{
		valuer.Identity(), valuer.Identity(), valuer.Identity(),
	}

	tests := []struct {
		name  string
		it    NodeIterator
		nodes []Node
	}{
		{"nil", nil, []Node{}},
		{"not nil", NodeSlice(nodes), nodes},
	}

	for _, tt := range tests {
		testNodeCollection(t, &nodeCollection{tt.it}, tt.nodes, false)
	}
}

func TestNodeSlice(t *testing.T) {
	tests := []struct {
		name  string
		nodes []Node
	}{
		{"nil", nil},
		{"0", []Node{valuer.Identity()}},
		{"1", []Node{valuer.Identity(), valuer.Identity()}},
		{"n", []Node{valuer.Identity(), valuer.Identity(), valuer.Identity(), valuer.Identity()}},
	}

	for _, tt := range tests {
		testNodeCollection(t, NodeSlice(tt.nodes), tt.nodes, false)
	}
}

func TestNodeSet(t *testing.T) {
	tests := []struct {
		name  string
		nodes []Node
	}{
		{"nil", nil},
		{"0", []Node{valuer.Identity()}},
		{"1", []Node{valuer.Identity(), valuer.Identity()}},
		{"n", []Node{valuer.Identity(), valuer.Identity(), valuer.Identity(), valuer.Identity()}},
	}

	for _, tt := range tests {
		set := newNodeSet(tt.nodes...)
		testNodeSet(t, set, tt.nodes, false)
	}

	t.Run("Add", func(t *testing.T) {
		set := newNodeSet()

		t.Run("Add nil", func(t *testing.T) {
			l := set.Len()
			set.Add(nil)
			l2 := set.Len()
			assert.Equal(t, l, l2)
		})

		t.Run("Add node", func(t *testing.T) {
			l := set.Len()
			set.Add(valuer.Identity())
			l2 := set.Len()
			assert.Equal(t, l+1, l2)
		})

		t.Run("Add node added before", func(t *testing.T) {
			node := valuer.Identity()
			set.Add(node)
			l := set.Len()
			set.Add(node)
			l2 := set.Len()
			assert.Equal(t, l, l2)
		})
	})

	t.Run("Remove", func(t *testing.T) {
		node := valuer.Identity()
		set := newNodeSet(node)

		t.Run("Remove nil", func(t *testing.T) {
			l := set.Len()
			set.Remove(nil)
			l2 := set.Len()
			assert.Equal(t, l, l2)
		})

		t.Run("Remove node", func(t *testing.T) {
			l := set.Len()
			set.Remove(node)
			l2 := set.Len()
			assert.Equal(t, l-1, l2)
		})

	})
}
