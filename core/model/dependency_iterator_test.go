package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	depArr := []Dependency{
		&dependency{rType: TypeOf(1)},
		&dependency{rType: TypeOf(1)},
		&dependency{rType: TypeOf(1)},
	}

	t.Run("to array", func(t *testing.T) {
		var arr []Dependency
		isContinue := ArrayDependencyIterator(depArr).Iterate(func(d Dependency) bool {
			arr = append(arr, d)
			return true
		})

		assert.True(t, isContinue)
		assert.Equal(t, depArr, arr)
	})

	t.Run("interrupt", func(t *testing.T) {
		var arr []Dependency
		isContinue := ArrayDependencyIterator(depArr).Iterate(func(d Dependency) bool {
			arr = append(arr, d)
			return len(arr) < 2
		})

		assert.False(t, isContinue)
		assert.Equal(t, depArr[0:2], arr)
	})

	testDependencyIteratorFormat(t, ArrayDependencyIterator(depArr))
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

func testDependencyIteratorFormat(t *testing.T, di DependencyIterator) {
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
		assert.Equal(t, "["+strings.Join(strArr, ", ")+"]", str)
		assert.Equal(t, "["+strings.Join(verboseStrArr, ", ")+"]", verboseStr)
	})
}
