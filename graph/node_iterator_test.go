package graph

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func nodesWithAttrsFromMustDistinct(ni NodeAndAttrsIterator) NodesWithAttrs {
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
		} else {
			panic(fmt.Sprintf("node %v has already met", node))
		}

		attrs.Iterate(func(key interface{}, value interface{}) bool {
			nodeAttrs.Set(key, value)
			return true
		})

		return true
	})

	return nodes
}

func Test_nodesWithAttrsFromMustDistinct(t *testing.T) {
	t.Run("all distinct nodes", func(t *testing.T) {
		nodes := NodesWithAttrs{
			1: {}, 2: {}, 3: {},
		}

		assert.Equal(t, nodes, nodesWithAttrsFromMustDistinct(nodes))
	})

	t.Run("with duplicate nodes", func(t *testing.T) {
		assert.Panics(t, func() {
			nodes := NodesWithAttrs{
				1: {}, 2: {}, 3: {},
			}
			it := combineNodeIterators(nodes, nodes)
			nodesWithAttrsFromMustDistinct(it)
		})
	})
}

func TestNodeAndAttrs_Iterate(t *testing.T) {
	tests := []struct {
		name string
		na   NodesWithAttrs
		want NodesWithAttrs
	}{
		{"NodesWithAttrs is nil", nil, NodesWithAttrs{}},
		{"NodesWithAttrs is empty", NodesWithAttrs{}, NodesWithAttrs{}},
		{"single value", NodesWithAttrs{1: Attrs{}}, NodesWithAttrs{1: Attrs{}}},
		{"multiple value", NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}},
			NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NodesWithAttrsFrom(tt.na)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestNodeAndAttrs_InterruptDuringIterate(t *testing.T) {
	t.Run("interrupt iteration", func(t *testing.T) {
		res := make(NodesWithAttrs)
		na := NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}, 2: Attrs{5: 6}, "b": Attrs{7: 8}}
		didNotInterrupted := na.Iterate(func(node Node, attrs AttrsView) bool {
			res[node] = AttrsFrom(attrs)
			return len(res) < 2
		})

		assert.Equal(t, 2, len(res))
		assert.False(t, didNotInterrupted)
	})
}

func TestNodeAndAttrs_UpdateAttrsDuringIterate(t *testing.T) {
	type args struct {
		key interface{}
		val interface{}
	}
	tests := []struct {
		name string
		na   NodesWithAttrs
		args args
		want NodesWithAttrs
	}{
		{"NodesWithAttrs is nil", nil, args{1, 2}, nil},
		{"NodesWithAttrs is empty", NodesWithAttrs{}, args{1, 2}, NodesWithAttrs{}},
		{"single value", NodesWithAttrs{1: Attrs{}}, args{1, 2},
			NodesWithAttrs{1: Attrs{1: 2}}},
		{"multiple value", NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}}, args{1, 3},
			NodesWithAttrs{1: Attrs{1: 3}, "a": Attrs{3: 4, 1: 3}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.na.Iterate(func(node Node, attrs AttrsView) bool {
				attrs.Set(tt.args.key, tt.args.val)
				return true
			})

			assert.Equal(t, tt.want, tt.na)
		})
	}

}

func Test_graphNodes_Iterate(t *testing.T) {
	type fields struct {
		graphNodes NodesWithAttrs
		nodes      []Node
	}
	tests := []struct {
		name   string
		fields fields
		want   NodesWithAttrs
	}{
		{"nodes is nil", fields{nil, nil}, NodesWithAttrs{}},
		{"nodes is nil 2", fields{nil, []Node{1, 2}}, NodesWithAttrs{}},
		{"nodes is empty", fields{NodesWithAttrs{}, nil}, NodesWithAttrs{}},
		{"single node", fields{NodesWithAttrs{1: Attrs{"a": "b"}}, nil},
			NodesWithAttrs{}},
		{"multiple nodes", fields{NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}}, nil},
			NodesWithAttrs{}},
		{"restricted nodes",
			fields{
				NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}},
				[]Node{"a"},
			},
			NodesWithAttrs{"a": Attrs{3: 4}},
		},
		{"restricted nodes does no exist",
			fields{
				NodesWithAttrs{1: Attrs{1: 2}, "a": Attrs{3: 4}},
				[]Node{"c"},
			},
			NodesWithAttrs{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nodesIter NodeIterator
			if tt.fields.nodes != nil {
				nodesIter = nodeIterateFunc(func(f func(node Node) bool) bool {
					for _, n := range tt.fields.nodes {
						if !f(n) {
							return false
						}
					}
					return true
				})
			}

			g := NewDirectedGraph()
			for node, attrs := range tt.fields.graphNodes {
				g.AddNodeWithAttrs(node, attrs)
			}

			ni := &graphNodes{
				graph: g,
				nodes: nodesIter,
			}
			res := NodesWithAttrsFrom(ni)
			assert.Equal(t, tt.want, res)
		})
	}

	t.Run("graph is nil", func(t *testing.T) {
		nodes := nodeIterateFunc(func(f func(node Node) bool) bool {
			for _, n := range []Node{1, 2, 3} {
				if !f(n) {
					return false
				}
			}
			return true
		})

		ni := &graphNodes{
			graph: nil,
			nodes: nodes,
		}
		res := NodesWithAttrsFrom(ni)
		assert.Equal(t, NodesWithAttrs{}, res)
	})
}

func TestNodesWithAttrsFrom(t *testing.T) {
	type args struct {
		na NodeAndAttrsIterator
	}
	tests := []struct {
		name string
		args args
		want NodesWithAttrs
	}{
		{"nil NodeAndAttrsIterator", args{nil}, NodesWithAttrs{}},
		{"single node", args{NodesWithAttrs{1: Attrs{"a": "b"}}}, NodesWithAttrs{1: Attrs{"a": "b"}}},
		{"multiple node", args{NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{"a": "b"}}},
			NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{"a": "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NodesWithAttrsFrom(tt.args.na), "NodesWithAttrsFrom(%v)", tt.args.na)
		})
	}
}

func Test_emptyNodeIterator_Iterate(t *testing.T) {
	assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(emptyNodeIterator{}))
}

func TestNodeEntry_Iterate(t *testing.T) {
	t.Run("iterate", func(t *testing.T) {
		entry := &NodeEntry{
			node:  1,
			attrs: Attrs{"a": "b"},
		}
		assert.Equal(t, NodesWithAttrs{1: Attrs{"a": "b"}}, NodesWithAttrsFrom(entry))
	})

	t.Run("entry is nil", func(t *testing.T) {
		var entry *NodeEntry
		assert.Equal(t, NodesWithAttrs{}, NodesWithAttrsFrom(entry))
	})
}

func Test_combinedNodeIterator_Iterate(t *testing.T) {
	type args struct {
		//left  NodeAndAttrsIterator
		//right NodeAndAttrsIterator

		it *combinedNodeIterator
	}
	tests := []struct {
		name string
		args args
		want NodesWithAttrs
	}{
		{"it is nil",
			args{nil},
			NodesWithAttrs{},
		},
		{"both args are nil",
			args{
				&combinedNodeIterator{nil, nil},
			},
			NodesWithAttrs{},
		},
		{"left is nil",
			args{
				&combinedNodeIterator{nil, NodesWithAttrs{1: Attrs{"a": "b"}}},
			},
			NodesWithAttrs{1: Attrs{"a": "b"}},
		},
		{"right is nil",
			args{
				&combinedNodeIterator{NodesWithAttrs{2: Attrs{"a": "b"}}, nil},
			},
			NodesWithAttrs{2: Attrs{"a": "b"}},
		},
		{"both args are not nil",
			args{
				&combinedNodeIterator{
					NodesWithAttrs{2: Attrs{"a": "b"}},
					NodesWithAttrs{1: Attrs{"a": "b"}},
				},
			},
			NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{"a": "b"}},
		},
		{"left and right are same",
			args{
				&combinedNodeIterator{
					NodesWithAttrs{2: Attrs{"a": "b"}},
					NodesWithAttrs{2: Attrs{"a": "b"}},
				},
			},
			NodesWithAttrs{2: Attrs{"a": "b"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NodesWithAttrsFrom(tt.args.it))
		})
	}

	t.Run("interrupt during iteration", func(t *testing.T) {
		t.Run("interrupt during left iterator iterating", func(t *testing.T) {
			left := NodesWithAttrs{1: {}, 2: {}, 3: {}}
			right := NodesWithAttrs{4: {}, 5: {}, 6: {}}
			it := &combinedNodeIterator{
				left:  left,
				right: right,
			}

			res := NodesWithAttrs{}

			it.Iterate(func(node Node, attrs AttrsView) bool {
				res[node] = AttrsFrom(attrs)
				return len(res) < 2
			})

			assert.Equal(t, 2, len(NodesWithAttrsFrom(res)))
		})

		t.Run("interrupt after left iterator end", func(t *testing.T) {
			left := NodesWithAttrs{1: {}, 2: {}, 3: {}}
			right := NodesWithAttrs{4: {}, 5: {}, 6: {}}
			it := &combinedNodeIterator{
				left:  left,
				right: right,
			}

			res := NodesWithAttrs{}

			it.Iterate(func(node Node, attrs AttrsView) bool {
				res[node] = AttrsFrom(attrs)
				return len(res) < 3
			})

			assert.Equal(t, left, NodesWithAttrsFrom(res))
		})

		t.Run("interrupt during right iterator iterating", func(t *testing.T) {
			left := NodesWithAttrs{1: {}, 2: {}, 3: {}}
			right := NodesWithAttrs{4: {}, 5: {}, 6: {}}
			it := &combinedNodeIterator{
				left:  left,
				right: right,
			}

			res := NodesWithAttrs{}

			it.Iterate(func(node Node, attrs AttrsView) bool {
				res[node] = AttrsFrom(attrs)
				return len(res) < 5
			})

			assert.Equal(t, 5, len(NodesWithAttrsFrom(res)))
		})

		t.Run("interrupt after right iterator iterate end", func(t *testing.T) {
			left := NodesWithAttrs{1: {}, 2: {}, 3: {}}
			right := NodesWithAttrs{4: {}, 5: {}, 6: {}}
			it := &combinedNodeIterator{
				left:  left,
				right: right,
			}

			res := NodesWithAttrs{}

			r := it.Iterate(func(node Node, attrs AttrsView) bool {
				res[node] = AttrsFrom(attrs)
				return len(res) < 6
			})

			assert.Equal(t, 6, len(NodesWithAttrsFrom(res)))
			assert.False(t, r)
		})
	})
}

func Test_combineNodeIterators(t *testing.T) {
	type args struct {
		iterators []NodeAndAttrsIterator
	}
	tests := []struct {
		name string
		args args
		want NodeAndAttrsIterator
	}{
		{"iterators is nil", args{nil}, NodesWithAttrs{}},
		{"iterators is empty", args{[]NodeAndAttrsIterator{}}, NodesWithAttrs{}},
		{"single iterator",
			args{[]NodeAndAttrsIterator{
				NodesWithAttrs{1: Attrs{"a": "b"}},
			}},
			NodesWithAttrs{1: Attrs{"a": "b"}},
		},
		{"several iterators",
			args{[]NodeAndAttrsIterator{
				NodesWithAttrs{1: Attrs{"a": "b"}}, NodesWithAttrs{2: Attrs{"a": "b"}}, NodesWithAttrs{3: Attrs{}},
				NodesWithAttrs{3: Attrs{}}, NodesWithAttrs{2: Attrs{"a": "b"}}, NodesWithAttrs{1: Attrs{"a": "b"}},
				nil,
			}},
			NodesWithAttrs{1: Attrs{"a": "b"}, 2: Attrs{"a": "b"}, 3: Attrs{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := combineNodeIterators(tt.args.iterators...)
			assert.Equal(t, tt.want, NodesWithAttrsFrom(it))
		})
	}
}

func Test_nodeIterateFunc_Iterate(t *testing.T) {
	tests := []struct {
		name string
		f    nodeIterateFunc
		want map[Node]struct{}
	}{
		{"function is nil", nil, map[Node]struct{}{}},
		{"function with nodes", func(f func(node Node) bool) bool {
			for _, n := range []Node{1, 2} {
				if !f(n) {
					return false
				}
			}
			return true
		}, map[Node]struct{}{1: {}, 2: {}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := map[Node]struct{}{}
			tt.f.Iterate(func(node Node) bool {
				nodes[node] = struct{}{}
				return true
			})

			assert.Equal(t, tt.want, nodes)
		})
	}

	t.Run("interrupt during iteration", func(t *testing.T) {
		it := nodeIterateFunc(func(f func(node Node) bool) bool {
			for _, n := range []Node{1, 2, 3, 4, 5} {
				if !f(n) {
					return false
				}
			}
			return true
		})

		nodes := map[Node]struct{}{}
		r := it.Iterate(func(node Node) bool {
			nodes[node] = struct{}{}
			return len(nodes) < 2
		})

		assert.False(t, r)
		assert.Equal(t, 2, len(nodes))
	})
}
