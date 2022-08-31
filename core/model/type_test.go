package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeOf(t *testing.T) {
	type testInterface interface{}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want reflect.Type
	}{
		{"reflect.Type", args{reflect.TypeOf(1)}, reflect.TypeOf(1)},
		{"reflect.Type of pointer of interface", args{reflect.TypeOf((*testInterface)(nil))},
			reflect.TypeOf((*testInterface)(nil)).Elem()},
		{"reflect.Value of not pointer", args{reflect.ValueOf(1)}, reflect.TypeOf(1)},
		{"reflect.Value of pointer of interface", args{reflect.ValueOf((*testInterface)(nil))},
			reflect.TypeOf((*testInterface)(nil)).Elem()},
		{"not pointer", args{1}, reflect.TypeOf(1)},
		{"pointer of interface", args{(*testInterface)(nil)},
			reflect.TypeOf((*testInterface)(nil)).Elem()},
		{"nil of interface", args{testInterface(nil)}, nil},
		{"*interface{}", args{(*interface{})(nil)}, reflect.TypeOf((*interface{})(nil)).Elem()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, TypeOf(tt.args.v), "TypeOf(%v)", tt.args.v)
		})
	}
}

func typesForTest() []reflect.Type {
	type testStruct struct{}
	type testInterface interface{}

	return []reflect.Type{
		TypeOf(1),
		TypeOf("a"),
		TypeOf('a'),
		TypeOf(int64(1)),
		TypeOf(testStruct{}),
		TypeOf(&testStruct{}),
		TypeOf((*testInterface)(nil)),
	}
}

func Test_typeSet_Has(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set *typeSet
		for _, rt := range typesForTest() {
			assert.False(t, set.Has(rt))
		}
	})

	t.Run("before and after adding type", func(t *testing.T) {
		set := newTypeSet()

		for _, rt := range typesForTest() {
			assert.False(t, set.Has(rt))
			set.Add(rt)
			assert.True(t, set.Has(rt))
		}
	})

	t.Run("before and after deleting type", func(t *testing.T) {
		set := newTypeSet()

		for _, rt := range typesForTest() {
			set.Add(rt)
		}

		set.Iterate(func(rt reflect.Type) bool {
			assert.True(t, set.Has(rt))
			set.Del(rt)
			assert.False(t, set.Has(rt))
			return true
		})
	})
}

func Test_typeSet_Len(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set *typeSet
		assert.Equal(t, 0, set.Len())
	})

	t.Run("before and after adding type", func(t *testing.T) {
		set := newTypeSet()
		for i, rt := range typesForTest() {
			assert.Equal(t, i, set.Len())
			set.Add(rt)
			assert.Equal(t, i+1, set.Len())
		}
	})

	t.Run("add type already in set", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest() {
			set.Add(rt)
		}

		set.Iterate(func(rt reflect.Type) bool {
			l := set.Len()
			set.Add(rt)
			assert.Equal(t, l, set.Len())
			return true
		})
	})

	t.Run("before add after deleting type", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest() {
			set.Add(rt)
		}

		set.Iterate(func(rt reflect.Type) bool {
			l := set.Len()
			set.Del(rt)
			assert.Equal(t, l-1, set.Len())
			return true
		})

		assert.Equal(t, 0, set.Len())
	})
}

func Test_typeSet_Iterate(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set *typeSet
		r := set.Iterate(func(s reflect.Type) bool {
			t.Fail()
			return true
		})
		assert.True(t, r)
	})

	t.Run("iterate all", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest() {
			set.Add(rt)
		}

		types := make(map[reflect.Type]struct{})
		set.Iterate(func(rt reflect.Type) bool {
			types[rt] = struct{}{}
			return true
		})

		assert.Equal(t, len(typesForTest()), len(types))
	})

	t.Run("interrupt iteration", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest() {
			set.Add(rt)
		}

		types := make(map[reflect.Type]struct{})
		set.Iterate(func(rt reflect.Type) bool {
			types[rt] = struct{}{}

			return len(types) < 5
		})

		assert.Equal(t, 5, len(types))
	})
}

func Test_typeSet_Equal(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set1 *typeSet
		var set2 *typeSet
		assert.True(t, set1.Equal(set2))
	})

	t.Run("equal", func(t *testing.T) {
		set1 := newTypeSet()
		set2 := newTypeSet()
		for _, rt := range typesForTest() {
			set1.Add(rt)
			set2.Add(rt)
			assert.True(t, set1.Equal(set2))
		}

		for _, rt := range typesForTest() {
			set1.Del(rt)
			set2.Del(rt)
			assert.True(t, set1.Equal(set2))
		}
	})

	t.Run("not equal", func(t *testing.T) {
		set1 := newTypeSet()
		set2 := newTypeSet()
		set1.Add(TypeOf(0))
		set2.Add(TypeOf(""))
		assert.False(t, set1.Equal(set2))
	})

	t.Run("nil equal empty", func(t *testing.T) {
		var set1 *typeSet
		set2 := newTypeSet()
		assert.True(t, set1.Equal(set2))
	})
}

func Test_typeSet_Add(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set *typeSet
		for _, rt := range typesForTest() {
			assert.False(t, set.Has(rt))
			set.Add(rt)
			assert.False(t, set.Has(rt))
		}
	})

	t.Run("add types", func(t *testing.T) {
		set := newTypeSet()
		for i, rt := range typesForTest() {
			set.Add(rt)
			assert.True(t, set.Has(rt))
			assert.Equal(t, i+1, set.Len())
		}
	})

	t.Run("add type that already in set", func(t *testing.T) {
		types := typesForTest()

		set := newTypeSet(types...)

		set.Add(types[1])
		assert.Equal(t, types, set.types())
	})

	t.Run("add nil type", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest() {
			set.Add(rt)
		}
		oldTypes := set.types()

		set.Add(nil)
		assert.False(t, set.Has(nil))
		assert.Equal(t, len(typesForTest()), set.Len())
		assert.Equal(t, oldTypes, set.types())
	})
}

func Test_typeSet_Del(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set *typeSet
		for _, rt := range typesForTest() {
			assert.False(t, set.Has(rt))
			set.Del(rt)
			assert.False(t, set.Has(rt))
		}
	})

	t.Run("delete types", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest() {
			set.Add(rt)
		}

		set.Iterate(func(rt reflect.Type) bool {
			l := set.Len()
			set.Del(rt)
			assert.Equal(t, l-1, set.Len())
			assert.False(t, set.Has(rt))
			return true
		})
	})
}

func Test_typeSet_Format(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var set *typeSet
		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{}", s)
		vs := fmt.Sprintf("%+v", set)
		assert.Equal(t, "{}", vs)
	})

	t.Run("empty", func(t *testing.T) {
		set := newTypeSet()
		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{}", s)
		vs := fmt.Sprintf("%+v", set)
		assert.Equal(t, "{}", vs)
	})

	t.Run("1 type", func(t *testing.T) {
		set := newTypeSet()
		set.Add(TypeOf(1))

		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{int}", s)
		vs := fmt.Sprintf("%+v", set)
		assert.Equal(t, "{int}", vs)
	})

	t.Run("n type", func(t *testing.T) {
		set := newTypeSet()
		for _, rt := range typesForTest()[0:2] {
			set.Add(rt)
		}

		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{int, string}", s)
		vs := fmt.Sprintf("%+v", set)
		vsExpected := "{" +
			"int" +
			", " +
			"string" +
			"}"
		assert.Equal(t, vsExpected, vs)
	})
}

func Test_typeSet_clone(t *testing.T) {
	t.Run("typeSet is nil", func(t *testing.T) {
		var set *typeSet
		set2 := set.clone()
		assert.Nil(t, set2)
	})

	t.Run("equality", func(t *testing.T) {
		t1 := TypeOf(1)
		t2 := TypeOf("a")
		t3 := TypeOf('a')
		set := newTypeSet(t1, t2, t3)
		set2 := set.clone()

		assert.Equal(t, set.types(), set2.types())
		assert.Equal(t, set.indexByType, set2.indexByType)
		assert.Equal(t, set.lastIndex, set2.lastIndex)
	})

	t.Run("update isolation", func(t *testing.T) {
		t1 := TypeOf(1)
		t2 := TypeOf("a")
		t3 := TypeOf('a')
		t4 := TypeOf(int64(1))
		set := newTypeSet(t1, t2, t3)
		set2 := set.clone()

		set2.Add(t4)
		assert.Equal(t, []reflect.Type{t1, t2, t3}, set.types())
		assert.Equal(t, 3, set.lastIndex)
		assert.Equal(t, map[reflect.Type]int{t1: 0, t2: 1, t3: 2}, set.indexByType)

		set3 := set.clone()
		set.Add(t4)

		assert.Equal(t, []reflect.Type{t1, t2, t3}, set3.types())
		assert.Equal(t, 3, set3.lastIndex)
		assert.Equal(t, map[reflect.Type]int{t1: 0, t2: 1, t3: 2}, set3.indexByType)
	})
}

func Test_newTypeSet(t *testing.T) {
	t.Run("empty types", func(t *testing.T) {
		set := newTypeSet()

		assert.Equal(t, 0, set.Len())
	})

	t.Run("one type", func(t *testing.T) {
		set := newTypeSet(TypeOf(1))

		assert.Equal(t, 1, set.Len())
		assert.True(t, set.Has(TypeOf(1)))
	})

	t.Run("multiple types", func(t *testing.T) {
		set := newTypeSet(typesForTest()...)

		assert.Equal(t, len(typesForTest()), set.Len())
		for _, rt := range typesForTest() {
			set.Has(rt)
		}
	})
}
