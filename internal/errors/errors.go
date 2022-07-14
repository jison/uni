package errors

import (
	"fmt"
	"strings"
)

type multiError struct {
	errors []error
}

func (me *multiError) Error() string {
	var b strings.Builder
	for _, err := range me.errors {
		b.WriteString(err.Error())
		b.WriteString("\n")
	}
	return b.String()
}

func New(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func Bug(err error) error {
	return fmt.Errorf("looks like you have found a bug in uni. %w", err)
}

func Bugf(format string, a ...interface{}) error {
	return Bug(fmt.Errorf(format, a...))
}

func Merge(errors ...error) error {
	errs := make([]error, 0, len(errors))
	for _, e := range errors {
		if e == nil {
			continue
		}
		errs = append(errs, e)
	}
	if len(errs) == 0 {
		return nil
	} else if len(errs) == 1 {
		return errs[0]
	}

	return &multiError{errors: errs}
}
