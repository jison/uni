package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testDependencyIterator(t *testing.T, it DependencyIterator, want []Dependency) {
	t.Run("iterate", func(t *testing.T) {
		m1 := map[Dependency]struct{}{}
		m2 := map[Dependency]struct{}{}

		for _, i := range want {
			m1[i] = struct{}{}
		}

		r := it.Iterate(func(i Dependency) bool {
			m2[i] = struct{}{}
			return true
		})

		assert.True(t, r)
		assert.Equal(t, m1, m2)
	})

	t.Run("interrupt", func(t *testing.T) {
		if len(want) == 0 {
			n := 0
			r := it.Iterate(func(_ Dependency) bool {
				n += 1
				return false
			})
			assert.True(t, r)
			assert.Equal(t, 0, n)
		} else {
			var half []Dependency
			r := it.Iterate(func(i Dependency) bool {
				half = append(half, i)
				return len(half) < len(want)/2
			})

			assert.False(t, r)

			expected := len(want) / 2
			if expected == 0 {
				expected = 1
			}
			assert.Equal(t, expected, len(half))
		}
	})

	t.Run("format", func(t *testing.T) {
		testDependencyIteratorFormat(t, it)
	})
}

func testDependencyIteratorFormat(t *testing.T, di DependencyIterator) {
	getAllPermutations := func(arr []string) [][]string {
		var res [][]string
		l := len(arr)
		var backtrack func(int)
		backtrack = func(first int) {
			if first == l {
				tmp := make([]string, l)
				copy(tmp, arr)
				res = append(res, tmp)
			}
			for i := first; i < l; i++ {
				arr[first], arr[i] = arr[i], arr[first]
				backtrack(first + 1)
				arr[first], arr[i] = arr[i], arr[first]
			}
		}

		backtrack(0)
		return res
	}

	getAllExpected := func(arr []string) []string {
		var res []string
		permutations := getAllPermutations(arr)
		for _, p := range permutations {
			res = append(res, "["+strings.Join(p, ", ")+"]")
		}
		return res
	}

	t.Run("Format", func(t *testing.T) {
		var strArr []string
		var verboseStrArr []string
		di.Iterate(func(d Dependency) bool {
			strArr = append(strArr, fmt.Sprintf("%v", d))
			verboseStrArr = append(verboseStrArr, fmt.Sprintf("%+v", d))
			return true
		})

		str := fmt.Sprintf("%v", di)
		verboseStr := fmt.Sprintf("%+v", di)

		assert.Contains(t, getAllExpected(strArr), str)
		assert.Contains(t, getAllExpected(verboseStrArr), verboseStr)
		//assert.Equal(t, "["+strings.Join(strArr, ", ")+"]", str)
		//assert.Equal(t, "["+strings.Join(verboseStrArr, ", ")+"]", verboseStr)
	})
}

func Test_emptyDependencyIterator_Iterate(t *testing.T) {
	t.Run("Iterate", func(t *testing.T) {
		it := emptyDependencyIterator{}

		isContinue := it.Iterate(func(d Dependency) bool {
			assert.Fail(t, "should no reach here")
			return true
		})
		assert.True(t, isContinue)

		isContinue2 := it.Iterate(func(d Dependency) bool {
			assert.Fail(t, "should no reach here")
			return false
		})
		assert.True(t, isContinue2)
	})

	testDependencyIteratorFormat(t, emptyDependencyIterator{})
}

func TestArrayDependencyIterator_Iterate(t *testing.T) {
	d1 := &dependency{rType: TypeOf(1)}
	d2 := &dependency{rType: TypeOf(1)}
	d3 := &dependency{rType: TypeOf(1)}

	tests := []struct {
		name string
		deps []Dependency
	}{
		{"nil", nil},
		{"0", []Dependency{}},
		{"1", []Dependency{d1}},
		{"2", []Dependency{d1, d2}},
		{"n", []Dependency{d1, d2, d3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := ArrayDependencyIterator(tt.deps)
			testDependencyIterator(t, it, tt.deps)
		})
	}
}

func dependencyIteratorToArray(di DependencyIterator) []Dependency {
	var arr []Dependency
	di.Iterate(func(d Dependency) bool {
		arr = append(arr, d)
		return true
	})
	return arr
}

func Test_combinedDependencyIterator_Iterate(t *testing.T) {
	depIt := ArrayDependencyIterator{
		&dependency{rType: TypeOf(1)},
		&dependency{rType: TypeOf(1)},
		&dependency{rType: TypeOf(1)},
	}

	depIt2 := ArrayDependencyIterator{
		&dependency{rType: TypeOf(1)},
		&dependency{rType: TypeOf(1)},
		&dependency{rType: TypeOf(1)},
	}

	t.Run("left is empty", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  emptyDependencyIterator{},
			right: depIt,
		}

		assert.Equal(t, dependencyIteratorToArray(depIt), dependencyIteratorToArray(cdi))
	})

	t.Run("right is empty", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  depIt,
			right: emptyDependencyIterator{},
		}

		assert.Equal(t, dependencyIteratorToArray(depIt), dependencyIteratorToArray(cdi))
	})

	t.Run("left and right is empty", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  emptyDependencyIterator{},
			right: emptyDependencyIterator{},
		}

		assert.Equal(t, []Dependency(nil), dependencyIteratorToArray(cdi))
	})

	t.Run("left and right is not empty", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  depIt,
			right: depIt2,
		}

		var allDep []Dependency
		allDep = append(allDep, ([]Dependency)(depIt)...)
		allDep = append(allDep, ([]Dependency)(depIt2)...)

		assert.Equal(t, allDep, dependencyIteratorToArray(cdi))
	})

	t.Run("interrupt on left", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  depIt,
			right: depIt2,
		}

		var allDep []Dependency
		allDep = append(allDep, ([]Dependency)(depIt)...)
		allDep = append(allDep, ([]Dependency)(depIt2)...)

		var arr []Dependency
		cdi.Iterate(func(d Dependency) bool {
			arr = append(arr, d)
			return len(arr) < 2
		})

		assert.Equal(t, allDep[0:2], arr)
	})

	t.Run("interrupt at right", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  depIt,
			right: depIt2,
		}

		var allDep []Dependency
		allDep = append(allDep, ([]Dependency)(depIt)...)
		allDep = append(allDep, ([]Dependency)(depIt2)...)

		var arr []Dependency
		cdi.Iterate(func(d Dependency) bool {
			arr = append(arr, d)
			return len(arr) < 5
		})

		assert.Equal(t, allDep[0:5], arr)
	})

	t.Run("format", func(t *testing.T) {
		cdi := &combinedDependencyIterator{
			left:  depIt,
			right: depIt2,
		}
		testDependencyIteratorFormat(t, cdi)
	})
}

func TestCombineDependencyIterators(t *testing.T) {
	var its [][]Dependency
	n := 10
	for i := 0; i < n; i++ {
		its = append(its, []Dependency{
			&dependency{rType: TypeOf(1)},
			&dependency{rType: TypeOf(1)},
		})
	}

	type args struct {
		its [][]Dependency
	}
	tests := []struct {
		name string
		args args
	}{
		{"0", args{nil}},
		{"1", args{its[0:1]}},
		{"2", args{its[0:2]}},
		{"n", args{its}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var allDep []Dependency
			var allIt []DependencyIterator
			for _, it := range tt.args.its {
				allDep = append(allDep, it...)
				allIt = append(allIt, ArrayDependencyIterator(it))
			}

			di := CombineDependencyIterators(allIt...)

			assert.Equal(t, allDep, dependencyIteratorToArray(di))
		})
	}
}
