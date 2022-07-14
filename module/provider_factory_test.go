package module

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFromNormalFunc(t *testing.T) {
	p, err := OfFunc(func(a int, b string) bool {
		return false
	})
	assert.Nil(t, err)
	assert.True(t, p.ParamAt(0).Type() == reflect.TypeOf(0))
	assert.True(t, p.ParamAt(1).Type() == reflect.TypeOf(""))
	assert.True(t, p.ComponentAt(0).Type() == reflect.TypeOf(false))
}

func TestBuildFromFuncWithEmptyParam(t *testing.T) {
	p, err := OfFunc(func() bool {
		return false
	})
	assert.Nil(t, err)
	assert.True(t, p.ComponentAt(0).Type() == reflect.TypeOf(false))
}

func TestBuildFromFuncWithVariadic(t *testing.T) {
	p, err := OfFunc(func(a int, b ...string) bool {
		return false
	})
	assert.Nil(t, err)
	assert.True(t, p.ParamAt(0).Type() == reflect.TypeOf(0))
	assert.True(t, p.ParamAt(1).Type() == reflect.TypeOf([]string{}))
	assert.True(t, p.ComponentAt(0).Type() == reflect.TypeOf(false))
}

func TestBuildFromFuncWithError(t *testing.T) {
	p, err := OfFunc(func(a int) (bool, error) {
		return false, nil
	})
	assert.Nil(t, err)
	p.Param(0, func(b ParameterBuilder) {

	})
	assert.True(t, p.ParamAt(0).Type() == reflect.TypeOf(0))
	assert.True(t, p.ComponentAt(0).Type() == reflect.TypeOf(false))
}

func TestBuildFromFuncWithErrors(t *testing.T) {
	p, err := OfFunc(func() (bool, error, int, error) {
		return false, nil, 0, nil
	})
	assert.Nil(t, err)
	assert.True(t, p.ComponentAt(0).Type() == reflect.TypeOf(false))
	assert.True(t, p.ComponentAt(1).Type() == reflect.TypeOf(1))
}

func TestBuildFromFuncWithEmptyResult(t *testing.T) {
	var err error
	_, err = OfFunc(func() {})
	assert.NotNil(t, err)

	_, err = OfFunc(func() error { return nil })
	assert.NotNil(t, err)
}
