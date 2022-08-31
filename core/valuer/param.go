package valuer

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type paramValuer struct {
	index int
}

type funcParam struct {
	index int
	val   reflect.Value
}

func (v *paramValuer) ValueOne(input Value) Value {
	if _, ok := input.AsError(); ok {
		return input
	}

	rVal, ok := input.AsSingle()
	if !ok {
		return ErrorValue(errors.Bugf("input value of param should be single"))
	}

	elem := funcParam{
		index: v.index,
		val:   rVal,
	}

	return SingleValue(reflect.ValueOf(elem))
}

func (v *paramValuer) String() string {
	return fmt.Sprintf("Param: %v", v.index)
}

func (v *paramValuer) Clone() OneInputValuer {
	return &paramValuer{v.index}
}

func (v *paramValuer) Equal(other interface{}) bool {
	o, ok := other.(*paramValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.index == o.index
}

func Param(index int) Valuer {
	return &oneInputValuer{&paramValuer{index}}
}
