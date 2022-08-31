package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jison/uni/core/model"
	"github.com/jison/uni/core/valuer"
	"github.com/stretchr/testify/assert"
)

func nodeIteratorToArray(ni NodeIterator) []Node {
	var nodes []Node
	ni.Iterate(func(node Node) bool {
		nodes = append(nodes, node)
		return true
	})
	return nodes
}

func Test_pathNode_Graph(t *testing.T) {
	t.Run("path is nil", func(t *testing.T) {
		var path *pathNode
		assert.Nil(t, path.Graph())
	})

	t.Run("graph is nil", func(t *testing.T) {
		path := NewPath(nil)
		assert.Nil(t, path.Graph())
	})

	t.Run("graph is not nil", func(t *testing.T) {
		rep := model.NewRepository(model.EmptyComponents())
		g := newDependenceGraph(rep)

		path := NewPath(g)
		assert.Equal(t, g, path.Graph())
	})
}

func Test_pathNode_Nodes(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var path *pathNode
		assert.Equal(t, []Node(nil), nodeIteratorToArray(path.Nodes()))
	})

	t.Run("0", func(t *testing.T) {
		path := NewPath(nil)
		assert.Equal(t, []Node(nil), nodeIteratorToArray(path.Nodes()))
	})

	t.Run("1", func(t *testing.T) {
		path := NewPath(nil)
		node := valuer.Identity()
		path = path.Append(node)
		assert.Equal(t, []Node{node}, nodeIteratorToArray(path.Nodes()))
	})

	t.Run("n", func(t *testing.T) {
		n := 10
		path := NewPath(nil)
		var nodes []Node
		for i := 0; i < n; i++ {
			node := valuer.Index(i)
			nodes = append(nodes, node)
			path = path.Append(node)
		}

		// reserve nodes
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}

		assert.Equal(t, nodes, nodeIteratorToArray(path.Nodes()))
	})
}

func Test_pathNode_Len(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var path *pathNode
		assert.Equal(t, 0, path.Len())
	})

	t.Run("0", func(t *testing.T) {
		path := NewPath(nil)
		assert.Equal(t, 0, path.Len())
	})

	t.Run("1", func(t *testing.T) {
		path := NewPath(nil)
		node := valuer.Identity()
		path = path.Append(node)
		assert.Equal(t, 1, path.Len())
	})

	t.Run("n", func(t *testing.T) {
		n := 10
		path := NewPath(nil)
		for i := 0; i < n; i++ {
			node := valuer.Index(i)
			path = path.Append(node)
		}

		assert.Equal(t, n, path.Len())
	})
}

func Test_pathNode_Contains(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var path *pathNode
		node := valuer.Identity()
		assert.False(t, path.Contains(node))
	})

	t.Run("0", func(t *testing.T) {
		path := NewPath(nil)
		node := valuer.Identity()
		assert.False(t, path.Contains(node))
	})

	t.Run("1", func(t *testing.T) {
		path := NewPath(nil)
		node := valuer.Identity()
		path = path.Append(node)
		assert.True(t, path.Contains(node))
	})

	t.Run("n", func(t *testing.T) {
		n := 10
		path := NewPath(nil)
		var nodes []Node
		for i := 0; i < n; i++ {
			node := valuer.Index(i)
			nodes = append(nodes, node)
			path = path.Append(node)
		}

		for _, node := range nodes {
			assert.True(t, path.Contains(node))
		}
	})
}

func Test_pathNode_Append(t *testing.T) {
	t.Run("path is nil", func(t *testing.T) {
		var path *pathNode
		node := valuer.Identity()
		path2 := path.Append(node)
		assert.NotEqual(t, path, path2)
		assert.Equal(t, 1, path2.Len())
		assert.Equal(t, []Node{node}, nodeIteratorToArray(path2.Nodes()))
		assert.True(t, path2.Contains(node))
		assert.Nil(t, path2.Graph())
	})

	t.Run("graph is nil", func(t *testing.T) {
		path := NewPath(nil)
		node := valuer.Identity()
		path2 := path.Append(node)
		assert.NotEqual(t, path, path2)
		assert.Equal(t, 1, path2.Len())
		assert.Equal(t, []Node{node}, nodeIteratorToArray(path2.Nodes()))
		assert.True(t, path2.Contains(node))
		assert.Nil(t, path2.Graph())
	})

	t.Run("graph is not nil", func(t *testing.T) {
		rep := model.NewRepository(model.EmptyComponents())
		g := newDependenceGraph(rep)

		n := 10
		path := NewPath(g)
		var nodes []Node
		for i := 0; i < n; i++ {
			node := valuer.Index(i)
			nodes = append(nodes, node)
			path = path.Append(node)
		}

		// reserve nodes
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}

		assert.Equal(t, n, path.Len())
		assert.Equal(t, nodes, nodeIteratorToArray(path.Nodes()))
		assert.Equal(t, g, path.Graph())
	})
}

func Test_pathNode_Reversed(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var path *pathNode
		reversed := path.Reversed()
		assert.Equal(t, []Node(nil), nodeIteratorToArray(reversed.Nodes()))
		assert.Equal(t, path.Graph(), reversed.Graph())
		assert.Equal(t, path.Len(), reversed.Len())
	})

	t.Run("0", func(t *testing.T) {
		path := NewPath(nil)
		reversed := path.Reversed()
		assert.Equal(t, []Node(nil), nodeIteratorToArray(reversed.Nodes()))
		assert.Equal(t, path.Graph(), reversed.Graph())
		assert.Equal(t, path.Len(), reversed.Len())
	})

	t.Run("1", func(t *testing.T) {
		path := NewPath(nil)
		node := valuer.Identity()
		path = path.Append(node)
		reversed := path.Reversed()
		assert.Equal(t, []Node{node}, nodeIteratorToArray(reversed.Nodes()))
		assert.Equal(t, path.Graph(), reversed.Graph())
		assert.Equal(t, path.Len(), reversed.Len())
	})

	t.Run("n", func(t *testing.T) {
		n := 10
		path := NewPath(nil)
		var nodes []Node
		for i := 0; i < n; i++ {
			node := valuer.Index(i)
			nodes = append(nodes, node)
			path = path.Append(node)
		}
		reversed := path.Reversed()
		assert.Equal(t, nodes, nodeIteratorToArray(reversed.Nodes()))
		assert.Equal(t, path.Graph(), reversed.Graph())
		assert.Equal(t, path.Len(), reversed.Len())
	})
}

func Test_pathNode_Iterate(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var path *pathNode
		meet := false
		path.Iterate(func(node Node) bool {
			meet = true
			return true
		})

		assert.False(t, meet)
	})

	t.Run("0", func(t *testing.T) {
		path := &pathNode{}
		meet := false
		path.Iterate(func(node Node) bool {
			meet = true
			return true
		})

		assert.False(t, meet)
	})

	t.Run("1", func(t *testing.T) {
		path := NewPath(nil).Append(valuer.Identity()).(*pathNode)

		count := 0
		path.Iterate(func(node Node) bool {
			count += 1
			return true
		})

		assert.Equal(t, 1, count)
	})

	t.Run("n", func(t *testing.T) {
		n := 10
		path := NewPath(nil)
		for i := 0; i < n; i++ {
			path = path.Append(valuer.Identity())
		}
		pNode := path.(*pathNode)
		count := 0
		pNode.Iterate(func(node Node) bool {
			count += 1
			return true
		})

		assert.Equal(t, n, count)
	})

	t.Run("interrupt", func(t *testing.T) {
		n := 10
		path := NewPath(nil)
		for i := 0; i < n; i++ {
			path = path.Append(valuer.Identity())
		}
		pNode := path.(*pathNode)
		count := 0
		pNode.Iterate(func(node Node) bool {
			count += 1
			return count < 5
		})

		assert.Equal(t, 5, count)
	})
}

func Test_pathNode_Format(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var path *pathNode
		str1 := fmt.Sprintf("%v", path)
		str2 := fmt.Sprintf("%+v", path)

		assert.Equal(t, "empty path", str1)
		assert.Equal(t, "empty path", str2)
	})

	t.Run("empty", func(t *testing.T) {
		rep := model.NewRepository(model.EmptyComponents())
		g := newDependenceGraph(rep)

		path := NewPath(g)
		str1 := fmt.Sprintf("%v", path)
		str2 := fmt.Sprintf("%+v", path)

		assert.Equal(t, "empty path", str1)
		assert.Equal(t, "empty path", str2)
	})

	t.Run("not empty", func(t *testing.T) {
		type testStruct struct {
			a int
		}

		m := model.NewModule(
			model.Struct(testStruct{}),
		)
		rep := model.NewRepository(m.AllComponents())
		g := newDependenceGraph(rep)
		path := NewPath(g)

		com := m.AllComponents().ToArray()[0]
		var node Node = com.Valuer()
		for {
			path = path.Append(node)
			r := g.InputNodesTo(node).Iterate(func(n Node) bool {
				node = n
				return false
			})
			if r {
				break
			}
		}
		var dep model.Dependency
		com.Provider().Dependencies().Iterate(func(d model.Dependency) bool {
			dep = d
			return false
		})

		t.Run("not verbose", func(t *testing.T) {
			expected := strings.Builder{}
			expected.WriteString("path:\n")
			expected.WriteString(fmt.Sprintf("\t%v\n", dep))
			expected.WriteString(fmt.Sprintf("\t%+v\n", com.Provider()))
			expected.WriteString(fmt.Sprintf("\t%v\n", com))

			str := fmt.Sprintf("%v", path)
			assert.Equal(t, expected.String(), str)
		})

		t.Run("verbose", func(t *testing.T) {
			expected := strings.Builder{}
			expected.WriteString("path:\n")
			expected.WriteString(fmt.Sprintf("\t%v\n", valuer.Error(&missingError{dep})))
			expected.WriteString(fmt.Sprintf("\t%v\n", dep))
			expected.WriteString(fmt.Sprintf("\t%+v\n", com.Provider()))
			expected.WriteString(fmt.Sprintf("\t%v\n", com))

			str := fmt.Sprintf("%+v", path)
			assert.Equal(t, expected.String(), str)
		})
	})
}

func TestNewPath(t *testing.T) {
	t.Run("graph is nil", func(t *testing.T) {
		path := NewPath(nil)
		assert.Nil(t, path.Graph())
	})

	t.Run("graph is not nil", func(t *testing.T) {
		rep := model.NewRepository(model.EmptyComponents())
		g := newDependenceGraph(rep)
		path := NewPath(g)
		assert.Equal(t, g, path.Graph())
	})
}
