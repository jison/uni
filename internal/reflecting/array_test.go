package reflecting

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayOfReflectValues(t *testing.T) {
	rVals := []reflect.Value{reflect.ValueOf(1), reflect.ValueOf("a")}
	vals, err := ArrayOfReflectValues(rVals)
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{1, "a"}, vals)
}

func TestReflectValuesOf(t *testing.T) {
	tests := []struct {
		args []interface{}
		want []reflect.Value
		err  bool
	}{
		{[]interface{}{1, "abc"}, []reflect.Value{reflect.ValueOf(1), reflect.ValueOf("abc")}, false},
		{[]interface{}{1, nil}, nil, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.args), func(t *testing.T) {
			vals, err := ReflectValuesOf(tt.args...)

			if !tt.err {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, vals)
			} else {
				assert.NotNil(t, err)
				assert.Nil(t, vals)
			}

		})
	}
}
