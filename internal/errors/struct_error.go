package errors

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//goland:noinspection SpellCheckingInspection
type StructError interface {
	error
	HasError() bool
	WithMain(error) StructError
	WithMainf(format string, a ...interface{}) StructError
	AddErrors(errs ...error) StructError
	AddErrorf(format string, a ...interface{}) StructError
}

type structError struct {
	mainError error
	subErrors []error
}

type errorCtx struct {
	parent    *errorCtx
	sb        *strings.Builder
	showIndex bool
	level     int
	index     int
}

func (ctx *errorCtx) newLevel() *errorCtx {
	return &errorCtx{
		parent: ctx, sb: ctx.sb, showIndex: ctx.showIndex, level: ctx.level + 1, index: 0,
	}
}

func (ctx *errorCtx) incrIndex() {
	ctx.index = ctx.index + 1
}

func (ctx *errorCtx) writeIndentN(n int) {
	for i := 0; i < n; i++ {
		ctx.writeString("\t")
	}
}

func (ctx *errorCtx) writeString(str string) {
	ctx.sb.WriteString(str)
}

func (ctx *errorCtx) writeSpace(n int) {
	for i := 0; i < n; i++ {
		ctx.sb.WriteString(" ")
	}
}

func (ctx *errorCtx) writeIndicator(toSpace bool) {
	var indexes []int
	cur := ctx
	for cur != nil && cur.showIndex {
		indexes = append(indexes, cur.index)
		cur = cur.parent
	}

	if len(indexes) == 0 {
		return
	}

	write := func(str string) {
		if !toSpace {
			ctx.writeString(str)
		} else {
			ctx.writeSpace(len(str))
		}
	}

	write("[")
	for i := len(indexes) - 1; i >= 0; i-- {
		if i < len(indexes)-1 {
			write(".")
		}
		write(strconv.Itoa(indexes[i]))
	}
	write("] ")
}

func (ctx *errorCtx) writeError(errStr string) {
	if ctx.level > 0 || ctx.index > 0 {
		ctx.writeString("\n")
	}

	lines := splitLines(errStr)
	for lineIdx, line := range lines {
		if lineIdx == 0 {
			ctx.writeIndentN(ctx.level)
			ctx.writeIndicator(false)
		} else {
			ctx.writeIndentN(ctx.level)
			ctx.writeIndicator(true)
		}
		ctx.writeString(line)
		if lineIdx < len(lines)-1 {
			ctx.writeString("\n")
		}
	}
}

func (ctx *errorCtx) string() string {
	//if ctx.index > 1 {
	//	return "[0] " + ctx.sb.String()
	//}
	return ctx.sb.String()
}

func (e *structError) LevelError(ctx *errorCtx) {
	subCtx := ctx
	if e.mainError != nil {
		ctx.writeError(e.mainError.Error())
		subCtx = ctx.newLevel()
	}

	if !subCtx.showIndex {
		subCtx.showIndex = len(e.subErrors) > 1
	}
	for _, subErr := range e.subErrors {
		if subStructErr, ok := subErr.(*structError); ok {
			subStructErr.LevelError(subCtx)
		} else {
			subCtx.writeError(subErr.Error())
			subCtx.incrIndex()
		}
	}

	if e.mainError != nil {
		ctx.incrIndex()
	}
}

func (e *structError) Error() string {
	var b strings.Builder
	ctx := &errorCtx{sb: &b}
	e.LevelError(ctx)
	return ctx.string()
}

func splitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func (e *structError) HasError() bool {
	if e == nil {
		return false
	}
	if e.mainError == nil && len(e.subErrors) == 0 {
		return false
	}
	return true
}

func (e *structError) WithMain(err error) StructError {
	if e == nil {
		if err == nil {
			return Empty()
		}
		return &structError{mainError: err, subErrors: nil}
	}
	if err == nil {
		if len(e.subErrors) == 0 {
			return Empty()
		}
	}
	return &structError{mainError: err, subErrors: e.subErrors}
}

func (e *structError) WithMainf(format string, a ...interface{}) StructError {
	main := fmt.Errorf(format, a...)
	return e.WithMain(main)
}

func (e *structError) AddErrors(errs ...error) StructError {
	var actualErrs []error
	for _, err := range errs {
		if err != nil {
			actualErrs = append(actualErrs, err)
		}
	}
	if len(actualErrs) == 0 {
		return e
	}

	if e == nil {
		return &structError{mainError: nil, subErrors: append([]error{}, actualErrs...)}
	}
	return &structError{mainError: e.mainError, subErrors: append(e.subErrors, actualErrs...)}
}

func (e *structError) AddErrorf(format string, a ...interface{}) StructError {
	err := fmt.Errorf(format, a...)
	return e.AddErrors(err)
}

func (e *structError) Is(err error) bool {
	if e2, ok := err.(*structError); ok {
		if errors.Is(e, e2.mainError) {
			return true
		}

		for _, subErr := range e2.subErrors {
			if errors.Is(e, subErr) {
				return true
			}
		}

		return false
	}

	if e.mainError != nil && errors.Is(e.mainError, err) {
		return true
	}

	for _, subErr := range e.subErrors {
		if errors.Is(subErr, err) {
			return true
		}
	}

	return false
}

func (e *structError) As(target interface{}) bool {
	if t, ok := target.(**structError); ok {
		*t = e
		return true
	}
	if t, ok := target.(*StructError); ok {
		*t = e
		return true
	}

	//goland:noinspection GoErrorsAs
	if errors.As(e.mainError, target) {
		return true
	}

	for _, err := range e.subErrors {
		//goland:noinspection GoErrorsAs
		if errors.As(err, target) {
			return true
		}
	}

	return false
}

func (e *structError) Unwrap() error {
	if e.mainError != nil {
		return e.mainError
	}

	for _, err := range e.subErrors {
		return err
	}

	return nil
}

func Empty() StructError {
	return (*structError)(nil)
}

func New(err error) StructError {
	var e *structError
	return e.WithMain(err)
}

//goland:noinspection SpellCheckingInspection
func Newf(format string, a ...interface{}) StructError {
	var e *structError
	return e.WithMainf(format, a...)
}

func Bug(err error) StructError {
	var e *structError
	return e.WithMainf("looks like you have found a bug in uni.").AddErrors(err)
}

//goland:noinspection SpellCheckingInspection
func Bugf(format string, a ...interface{}) StructError {
	return Bug(fmt.Errorf(format, a...))
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	//goland:noinspection GoErrorsAs
	return errors.As(err, target)
}
