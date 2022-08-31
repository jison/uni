package internal

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type structA struct {
	location string
	name     string
}

func (s structA) Name() string {
	return s.name
}

type structB struct {
	location string
	name     int
}

func (s structB) Name() string {
	return s.location
}

func TestTypeAssertion(t *testing.T) {
	var a interface{} = structA{"loc", "name"}
	aa, ok := a.(interface{ Name() string })
	assert.True(t, ok)
	fmt.Printf("%v\n", aa.Name())

}

func TestFuncParamArray(t *testing.T) {
	func1 := func(p []string) string {
		return ""
	}
	val := reflect.ValueOf(func1)
	fmt.Printf("%v \n", val.Type().In(0).Kind())
}

func TestVarargs(t *testing.T) {
	f1 := func(a ...interface{}) {
		fmt.Printf("%v\n", a)
	}

	f1(nil)
}

func TestReflectValueOfReflectValue(t *testing.T) {
	v := reflect.ValueOf(nil)
	vv := reflect.ValueOf(v)
	fmt.Printf("%v, %v\n", vv.IsValid(), vv)
}

func TestTypeOfAny(t *testing.T) {
	v := reflect.TypeOf((*interface{})(nil))
	v2 := v.Elem()
	fmt.Printf("%v, %v\n", v, v2)
}
