package valuer

import (
	"fmt"

	"github.com/jison/uni/internal/errors"
)

type errorValuer struct {
	err error
}

func (v *errorValuer) Value(_ []Value) Value {
	if v.err == nil {
		return ErrorValue(errors.Bugf("error can not be nil"))
	}
	return ErrorValue(v.err)
}

func (v *errorValuer) String() string {
	return fmt.Sprintf("Error: %v", v.err)
}

func (v *errorValuer) Clone() Valuer {
	return &errorValuer{v.err}
}

func (v *errorValuer) Equal(other interface{}) bool {
	o, ok := other.(*errorValuer)
	if !ok {
		return false
	}

	if v == nil || o == nil {
		return v == nil && o == nil
	}

	return v.err == o.err
}

func Error(err error) Valuer {
	return &errorValuer{err}
}
