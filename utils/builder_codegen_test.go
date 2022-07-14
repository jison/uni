package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenBuilder(t *testing.T) {
	targetType := reflect.TypeOf((*int)(nil)).Elem()
	src, err := GenBuilderForType(targetType)
	assert.Nil(t, err)

	fmt.Printf("%v", src)
	t.FailNow()
}

func TestSomething(t *testing.T) {
	s := "Abcde"
	fmt.Printf("%v\n", s[0:1])
	fmt.Printf("%v\n", s[1:2])
	t.FailNow()
}
