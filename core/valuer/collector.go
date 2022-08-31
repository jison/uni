package valuer

import (
	"fmt"
	"reflect"

	"github.com/jison/uni/internal/errors"
)

type collectorValuer struct {
	elementType reflect.Type
}

func (v *collectorValuer) Value(inputs []Value) Value {
	errs := errors.Empty()
	sliceType := reflect.SliceOf(v.elementType)
	newSlice := reflect.MakeSlice(sliceType, 0, len(inputs))
	for _, inVal := range inputs {
		if err, ok := inVal.AsError(); ok {
			errs = errs.AddErrors(err)
			continue
		}

		rVal, isSingle := inVal.AsSingle()
		if !isSingle {
			errs = errs.AddErrors(errors.Bugf("collect val should be single"))
			continue
		}

		if !rVal.Type().AssignableTo(v.elementType) {
			errs = errs.AddErrorf("%v (%v) is not assignable to %v", inVal, rVal.Type(), v.elementType)
			continue
		}
		if !rVal.CanInterface() {
			errs = errs.AddErrorf("%+v .CanInterface() is false", inVal)
			continue
		}
		newSlice = reflect.Append(newSlice, reflect.ValueOf(rVal.Interface()))
	}
	if errs.HasError() {
		return ErrorValue(errs)
	}

	return SingleValue(newSlice)
}

func (v *collectorValuer) String() string {
	return fmt.Sprintf("Collect: %+v", v.elementType)
}

func (v *collectorValuer) Clone() Valuer {
	return &collectorValuer{v.elementType}
}

func (v *collectorValuer) Equal(other interface{}) bool {
	o, ok := other.(*collectorValuer)
	if !ok {
		return false
	}
	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.elementType == o.elementType
}

func Collector(elementType reflect.Type) Valuer {
	return &collectorValuer{elementType}
}
