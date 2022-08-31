package valuer

import (
	"fmt"
	"github.com/jison/uni/internal/errors"
)

type Valuer interface {
	fmt.Stringer

	Value([]Value) Value
	Clone() Valuer
	Equal(interface{}) bool
}

type OneInputValuer interface {
	ValueOne(Value) Value
	String() string
	Clone() OneInputValuer
	Equal(interface{}) bool
}

type oneInputValuer struct {
	concreteValuer OneInputValuer
}

func (v *oneInputValuer) Value(inputs []Value) Value {
	if len(inputs) != 1 {
		err := errors.Bugf("%v can not apply to %v input Vertex",
			v.concreteValuer.String(), len(inputs))
		return ErrorValue(err)
	}

	val := inputs[0]
	if _, ok := val.AsError(); ok {
		return val
	}

	return v.concreteValuer.ValueOne(val)
}

func (v *oneInputValuer) String() string {
	return v.concreteValuer.String()
}

func (v *oneInputValuer) Clone() Valuer {
	return &oneInputValuer{v.concreteValuer.Clone()}
}

func (v *oneInputValuer) Equal(other interface{}) bool {
	o, ok := other.(*oneInputValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.concreteValuer.Equal(o.concreteValuer)
}
