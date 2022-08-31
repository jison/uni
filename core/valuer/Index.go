package valuer

import (
	"fmt"

	"github.com/jison/uni/internal/errors"
)

type indexValuer struct {
	index int
}

func (v *indexValuer) ValueOne(input Value) Value {
	arr, ok := input.AsArray()
	if !ok {
		return ErrorValue(errors.Bugf("index can not apply to non array value"))
	}

	if v.index < 0 || len(arr) <= v.index {
		return ErrorValue(errors.Bugf("index %v is out of range [0, %v].", v.index, len(arr)-1))
	}

	return SingleValue(arr[v.index])
}

func (v *indexValuer) String() string {
	return fmt.Sprintf("Index: %v", v.index)
}

func (v *indexValuer) Clone() OneInputValuer {
	return &indexValuer{index: v.index}
}

func (v *indexValuer) Equal(other interface{}) bool {
	o, ok := other.(*indexValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.index == o.index
}

func Index(index int) Valuer {
	return &oneInputValuer{&indexValuer{index}}
}
