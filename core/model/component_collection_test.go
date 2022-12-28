package model

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newComponentForTest() *component {
	return &component{}
}

func testComponentIterator(t *testing.T, it ComponentIterator, coms []Component) {
	t.Run("iterate", func(t *testing.T) {
		m1 := map[Component]struct{}{}
		m2 := map[Component]struct{}{}

		for _, c := range coms {
			m1[c] = struct{}{}
		}

		r := it.Iterate(func(com Component) bool {
			m2[com] = struct{}{}
			return true
		})

		assert.True(t, r)
		assert.Equal(t, m1, m2)
	})

	t.Run("interrupt", func(t *testing.T) {
		if len(coms) == 0 {
			n := 0
			r := it.Iterate(func(com Component) bool {
				n += 1
				return false
			})
			assert.True(t, r)
			assert.Equal(t, 0, n)
		} else {
			var half []Component
			r := it.Iterate(func(com Component) bool {
				half = append(half, com)
				return len(half) < len(coms)/2
			})

			assert.False(t, r)

			expected := len(coms) / 2
			if expected == 0 {
				expected = 1
			}
			assert.Equal(t, expected, len(half))
		}
	})
}

func componentIteratorToArray(it ComponentIterator) []Component {
	var arr []Component
	it.Iterate(func(c Component) bool {
		arr = append(arr, c)
		return true
	})
	return arr
}

func TestFuncComponentIterator(t *testing.T) {
	com1 := &component{}
	com2 := &component{}
	com3 := &component{}

	tests := []struct {
		name string
		coms []Component
		want []Component
	}{
		{"nil", nil, []Component{}},
		{"0", []Component{}, []Component{}},
		{"1", []Component{com1}, []Component{com1}},
		{"2", []Component{com1, com2}, []Component{com1, com2}},
		{"n", []Component{com1, com2, com3}, []Component{com1, com2, com3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := FuncComponentIterator(func(f func(Component) bool) bool {
				for _, c := range tt.coms {
					if !f(c) {
						return false
					}
				}
				return true
			})
			testComponentIterator(t, it, tt.want)
		})
	}
}

func Test_filteredComponentIterator(t *testing.T) {
	var comSeq []Component
	for i := 0; i < 10; i++ {
		c := newComponentForTest()
		c.SetName(strconv.Itoa(i))
		comSeq = append(comSeq, c)
	}

	arrIt := ComponentSlice(comSeq)
	predicate := func(c Component) bool {
		i, _ := strconv.Atoi(c.Name())
		return i%2 == 0
	}
	filteredIt := &filteredComponentIterator{arrIt, predicate}

	var filteredComSeq []Component
	for _, com := range comSeq {
		if predicate(com) {
			filteredComSeq = append(filteredComSeq, com)
		}
	}

	testComponentIterator(t, filteredIt, filteredComSeq)
}

func Test_combinedComponentIterator(t *testing.T) {
	comIt := ComponentSlice{
		newComponentForTest(), newComponentForTest(), newComponentForTest(),
	}

	comIt2 := ComponentSlice{
		newComponentForTest(), newComponentForTest(), newComponentForTest(),
	}

	t.Run("left is empty", func(t *testing.T) {
		cci := &combinedComponentIterator{
			left:  emptyComponentCollection{},
			right: comIt,
		}

		assert.Equal(t, componentIteratorToArray(comIt), componentIteratorToArray(cci))
	})

	t.Run("right is empty", func(t *testing.T) {
		cci := &combinedComponentIterator{
			left:  comIt,
			right: emptyComponentCollection{},
		}

		assert.Equal(t, componentIteratorToArray(comIt), componentIteratorToArray(cci))
	})

	t.Run("left and right is empty", func(t *testing.T) {
		cci := &combinedComponentIterator{
			left:  emptyComponentCollection{},
			right: emptyComponentCollection{},
		}

		assert.Equal(t, []Component(nil), componentIteratorToArray(cci))
	})

	t.Run("left and right is not empty", func(t *testing.T) {
		cci := &combinedComponentIterator{
			left:  comIt,
			right: comIt2,
		}

		var allDep []Component
		allDep = append(allDep, ([]Component)(comIt)...)
		allDep = append(allDep, ([]Component)(comIt2)...)

		assert.Equal(t, allDep, componentIteratorToArray(cci))
	})

	t.Run("interrupt on left", func(t *testing.T) {
		cdi := &combinedComponentIterator{
			left:  comIt,
			right: comIt2,
		}

		var allDep []Component
		allDep = append(allDep, ([]Component)(comIt)...)
		allDep = append(allDep, ([]Component)(comIt2)...)

		var arr []Component
		cdi.Iterate(func(d Component) bool {
			arr = append(arr, d)
			return len(arr) < 2
		})

		assert.Equal(t, allDep[0:2], arr)
	})

	t.Run("interrupt at right", func(t *testing.T) {
		cdi := &combinedComponentIterator{
			left:  comIt,
			right: comIt2,
		}

		var allDep []Component
		allDep = append(allDep, ([]Component)(comIt)...)
		allDep = append(allDep, ([]Component)(comIt2)...)

		var arr []Component
		cdi.Iterate(func(d Component) bool {
			arr = append(arr, d)
			return len(arr) < 5
		})

		assert.Equal(t, allDep[0:5], arr)
	})
}

func TestCombineComponents(t *testing.T) {
	n := 10
	var its []ComponentIterator
	for i := 0; i < n; i++ {
		var it ComponentSlice
		for ii := 0; ii < 10; ii++ {
			it = append(it, newComponentForTest())
		}
		its = append(its, it)
	}

	for _, i := range []int{0, 1, 2, n} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var coms []Component
			for _, it := range its[0:i] {
				coms = append(coms, ComponentsOfIterator(it).ToArray()...)
			}

			cc := CombineComponents(its[0:i]...)
			testComponentCollection(t, cc, coms, false)
		})
	}
}

func Test_distinctComponentIterator(t *testing.T) {
	t.Run("have same components", func(t *testing.T) {
		var cs ComponentSlice
		var coms []Component
		for i := 0; i < 10; i++ {
			c := newComponentForTest()
			cs = append(cs, c, c)
			coms = append(coms, c)
		}

		it := &distinctComponentIterator{cs}
		assert.Equal(t, len(coms), len(ComponentsOfIterator(it).ToArray()))

		testComponentIterator(t, it, coms)
	})

	t.Run("does not have same components", func(t *testing.T) {
		var cs ComponentSlice
		var coms []Component
		for i := 0; i < 10; i++ {
			c := newComponentForTest()
			cs = append(cs, c)
			coms = append(coms, c)
		}

		it := &distinctComponentIterator{cs}
		assert.Equal(t, len(coms), len(ComponentsOfIterator(it).ToArray()))
		testComponentIterator(t, it, coms)
	})
}

func testComponentCollection(t *testing.T, cc ComponentCollection, coms []Component, isRecursive bool) {
	testComponentIterator(t, cc, coms)

	t.Run("Each", func(t *testing.T) {
		m1 := map[Component]struct{}{}
		m2 := map[Component]struct{}{}

		for _, n := range coms {
			m1[n] = struct{}{}
		}

		cc.Each(func(c Component) {
			m2[c] = struct{}{}
		})

		assert.Equal(t, m1, m2)
	})

	t.Run("Filter", func(t *testing.T) {
		if isRecursive {
			return
		}
		set := map[Component]struct{}{}
		var filteredComs []Component
		for _, n := range coms {
			if rand.Intn(2) == 0 {
				continue
			}
			set[n] = struct{}{}
			filteredComs = append(filteredComs, n)
		}

		filteredNc := cc.Filter(func(c Component) bool {
			_, ok := set[c]
			return ok
		})
		testComponentCollection(t, filteredNc, filteredComs, true)
	})

	t.Run("ToArray", func(t *testing.T) {
		m1 := map[Component]struct{}{}
		m2 := map[Component]struct{}{}

		for _, c := range coms {
			m1[c] = struct{}{}
		}
		for _, c := range cc.ToArray() {
			m2[c] = struct{}{}
		}

		assert.Equal(t, m1, m2)
	})

	t.Run("ToSet", func(t *testing.T) {
		if isRecursive {
			return
		}
		testComponentSet(t, cc.ToSet(), coms, true)
	})

	t.Run("Distinct", func(t *testing.T) {
		m1 := map[Component]struct{}{}
		m2 := map[Component]struct{}{}

		cc.Distinct().Each(func(c Component) {
			m1[c] = struct{}{}
		})
		cc.ToSet().Each(func(c Component) {
			m2[c] = struct{}{}
		})

		assert.Equal(t, m1, m2)
	})

	t.Run("Format", func(t *testing.T) {
		test := func(format string) {
			var arr []string
			for _, c := range coms {
				arr = append(arr, fmt.Sprintf(format, c))
			}
			expected := "[" + strings.Join(arr, ", ") + "]"
			assert.Equal(t, expected, fmt.Sprintf(format, cc))
		}

		t.Run("not verbose", func(t *testing.T) {
			test("%v")
		})

		t.Run("verbose", func(t *testing.T) {
			test("%+v")
		})

		t.Run("#", func(t *testing.T) {
			test("%#v")
		})
	})
}

func testComponentSet(t *testing.T, cs ComponentSet, coms []Component, isRecursive bool) {
	testComponentCollection(t, cs, coms, isRecursive)

	t.Run("Contains", func(t *testing.T) {
		for _, n := range coms {
			assert.True(t, cs.Contains(n))
		}

		for i := 0; i < 10; i++ {
			c := newComponentForTest()
			assert.False(t, cs.Contains(c))
		}
	})

	t.Run("Len", func(t *testing.T) {
		assert.Equal(t, len(coms), cs.Len())
	})
}

func TestComponentSlice(t *testing.T) {
	var comSeq []Component
	for i := 0; i < 10; i++ {
		comSeq = append(comSeq, newComponentForTest())
	}

	testComponentCollection(t, ComponentSlice(comSeq), comSeq, false)
}

func TestComponentsOfIterator(t *testing.T) {
	var coms []Component
	for i := 0; i < 10; i++ {
		c := newComponentForTest()
		coms = append(coms, c)
	}
	it := ComponentSlice(coms)

	cs := ComponentsOfIterator(it)
	testComponentCollection(t, cs, coms, false)
}

func TestEmptyComponentCollection(t *testing.T) {
	cs := emptyComponentCollection{}
	testComponentCollection(t, cs, []Component(nil), false)
}

func Test_componentSet(t *testing.T) {
	var coms []Component
	for i := 0; i < 10; i++ {
		c := newComponentForTest()
		coms = append(coms, c)
	}

	cs := newComponentSet()

	t.Run("newComponentSet", func(t *testing.T) {
		c1 := newComponentForTest()
		c2 := newComponentForTest()
		set := newComponentSet(c1, c2)
		assert.True(t, set.Contains(c1))
		assert.True(t, set.Contains(c2))
		assert.Equal(t, 2, len(set.ToArray()))
	})

	t.Run("Add", func(t *testing.T) {
		for _, c := range coms {
			assert.False(t, cs.Contains(c))
			cs.Add(c)
			assert.True(t, cs.Contains(c))
		}
	})

	t.Run("Contains", func(t *testing.T) {
		for _, c := range coms {
			assert.True(t, cs.Contains(c))
			cs.Add(c)
			assert.True(t, cs.Contains(c))
			assert.Equal(t, len(cs.ToArray()), 10)
		}
	})

	testComponentSet(t, cs, coms, false)
}
