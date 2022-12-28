package reflecting

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsErrorType(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := errors.Newf("abc")
		assert.True(t, IsErrorType(reflect.TypeOf(err)))
	})

	t.Run("no error", func(t *testing.T) {
		assert.False(t, IsErrorType(reflect.TypeOf(1)))
	})

	t.Run("nil", func(t *testing.T) {
		assert.False(t, IsErrorType(nil))
	})
}

func TestIsErrorValue(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := errors.Newf("abc")
		assert.True(t, IsErrorValue(reflect.ValueOf(err)))
	})

	t.Run("invalid value", func(t *testing.T) {
		assert.False(t, IsErrorValue(reflect.ValueOf(nil)))
	})
}

func TestAsError(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := errors.Newf("abc")
		err2, ok := AsError(reflect.ValueOf(err))
		assert.True(t, ok)
		assert.Equal(t, err, err2)
	})

	t.Run("no error", func(t *testing.T) {
		_, ok := AsError(reflect.ValueOf("abc"))
		assert.False(t, ok)
	})

	t.Run("value invalid", func(t *testing.T) {
		_, ok := AsError(reflect.ValueOf(nil))
		assert.False(t, ok)
	})

	type testStruct struct {
		a error
	}

	t.Run("CanInterface() is false", func(t *testing.T) {
		ts := testStruct{a: errors.Newf("this is an error")}
		val := reflect.ValueOf(ts).Field(0)
		_, ok := AsError(val)
		assert.False(t, ok)
	})
}

func TestIsNilValue(t *testing.T) {
	var intPtr *int
	var intChan chan int
	var m map[int]string
	var slice []int
	var f func()
	var iface interface{}

	type args struct {
		val interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil", args{nil}, true},
		{"int", args{1}, false},
		{"nil ptr", args{intPtr}, true},
		{"ptr", args{&args{}}, false},
		{"nil intChan", args{intChan}, true},
		{"intChan", args{make(chan int)}, false},
		{"nil map", args{m}, true},
		{"map", args{map[interface{}]interface{}{1: 1}}, false},
		{"string", args{"a"}, false},
		{"empty array", args{[0]int{}}, false},
		{"array", args{[1]int{1}}, false},
		{"nil slice", args{slice}, true},
		{"slice", args{[]int{1}}, false},
		{"nil func", args{f}, true},
		{"func", args{func() {}}, false},
		{"interface{}", args{iface}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsNilValue(reflect.ValueOf(tt.args.val)))
		})
	}

}

func TestReflectValueOfArray(t *testing.T) {
	func1 := func(a, b int) int {
		return a + b
	}

	rFunc := ReflectFuncOfFunc(reflect.ValueOf(func1))
	callVals, callErr := rFunc([]reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2)})
	vals := make([]interface{}, 0, len(callVals))
	for _, v := range callVals {
		vals = append(vals, v.Interface())
	}

	assert.Nil(t, callErr)
	val := reflect.ValueOf(vals)
	assert.Equal(t, val.Index(0).Interface(), 3)
}

func TestIsKindOrPtrOfKind(t *testing.T) {
	str := ""
	strPtr := &str
	var nilStrPtr *string

	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(nilStrPtr), reflect.String))

	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(str), reflect.String))
	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(strPtr), reflect.String))

	assert.False(t, IsKindOrPtrOfKind(reflect.TypeOf(str), reflect.Ptr))
	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(strPtr), reflect.Ptr))

	assert.False(t, IsKindOrPtrOfKind(reflect.TypeOf(str), reflect.Int))
	assert.False(t, IsKindOrPtrOfKind(reflect.TypeOf(strPtr), reflect.Int))

	assert.False(t, IsKindOrPtrOfKind(nil, reflect.Interface))
}

func TestCanBeMapKey(t *testing.T) {
	type args struct {
		val interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"nil", args{nil}, true},
		{"int", args{1}, true},
		{"ptr", args{t}, true},
		{"string", args{"a"}, true},
		{"array", args{[1]int{1}}, true},
		{"slice", args{[]int{1}}, false},
		{"map", args{map[interface{}]interface{}{1: 1}}, false},
		{"func", args{func() {}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CanBeMapKey(tt.args.val), "CanBeMapKey(%v)", tt.args.val)
		})
	}
}
