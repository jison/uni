package reflecting

import (
	"reflect"
	"testing"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsErrorType(t *testing.T) {
	err := errors.Newf("abc")
	assert.True(t, IsErrorType(reflect.TypeOf(err)))

	assert.False(t, IsErrorType(reflect.TypeOf(1)))
}

func TestIsErrorValue(t *testing.T) {
	err := errors.Newf("abc")
	assert.True(t, IsErrorValue(reflect.ValueOf(err)))
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
}

func TestIsNilValue(t *testing.T) {
	val := reflect.ValueOf(nil)
	assert.True(t, IsNilValue(val))

	assert.False(t, IsNilValue(reflect.ValueOf(123)))
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

	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(str), reflect.String))
	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(strPtr), reflect.String))

	assert.False(t, IsKindOrPtrOfKind(reflect.TypeOf(str), reflect.Ptr))
	assert.True(t, IsKindOrPtrOfKind(reflect.TypeOf(strPtr), reflect.Ptr))

	assert.False(t, IsKindOrPtrOfKind(reflect.TypeOf(str), reflect.Int))
	assert.False(t, IsKindOrPtrOfKind(reflect.TypeOf(strPtr), reflect.Int))
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
