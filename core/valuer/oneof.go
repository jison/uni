package valuer

import (
	"fmt"
	"math/rand"

	"github.com/jison/uni/internal/errors"
)

type oneOfValuer struct {
	_ int // make every valuer different
}

func (v *oneOfValuer) Value(inputs []Value) Value {
	if len(inputs) == 0 {
		return ErrorValue(errors.Bugf("not values input for one of vertex"))
	}

	//errs := errors.Empty()
	//var availableVals []Value
	//for _, inVal := range inputs {
	//	if err, ok := inVal.AsError(); ok {
	//		errs = errs.AddErrors(err)
	//	} else {
	//		availableVals = append(availableVals, inVal)
	//	}
	//}
	//
	//if len(availableVals) == 0 {
	//	return ErrorValue(errs.WithMainf("all value of inputs are errors"))
	//}

	//valIndex := rand.Intn(len(availableVals))
	//return availableVals[valIndex]

	valIndex := rand.Intn(len(inputs))
	return inputs[valIndex]
}

func (v *oneOfValuer) String() string {
	return fmt.Sprintf("OneOf")
}

func (v *oneOfValuer) Clone() Valuer {
	return &oneOfValuer{}
}

func (v *oneOfValuer) Equal(other interface{}) bool {
	if v == nil && other == nil {
		return true
	}

	_, ok := other.(*oneOfValuer)
	return ok
}

func OneOf() Valuer {
	return &oneOfValuer{}
}
