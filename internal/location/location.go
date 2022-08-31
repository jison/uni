package location

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type Location interface {
	PkgName() string
	FuncName() string
	FullName() string
	FileName() string
	FileLine() int
}

type CallLocation interface {
	Location
	Callee() Location
}

type location struct {
	pkgName  string
	funcName string
	fileName string
	fileLine int
}

func (l *location) PkgName() string {
	return l.pkgName
}

func (l *location) FuncName() string {
	return l.funcName
}

func (l *location) FullName() string {
	return l.PkgName() + "." + l.FuncName()
}

func (l *location) FileName() string {
	return l.fileName
}

func (l *location) FileLine() int {
	return l.fileLine
}

func (l *location) Format(f fmt.State, verb rune) {
	if f.Flag('+') && verb == 'v' {
		_, _ = fmt.Fprintf(f, "%v.%v (%v:%d)", l.pkgName, l.funcName, l.fileName, l.fileLine)
	} else {
		_, _ = fmt.Fprintf(f, "%v:%d", l.fileName, l.fileLine)
	}
}

func (l *location) String() string {
	return fmt.Sprintf("%v", l)
}

type callLocation struct {
	location
	callee Location
}

func (cl *callLocation) Format(f fmt.State, verb rune) {
	if f.Flag('+') && verb == 'v' {
		_, _ = fmt.Fprintf(f, "call %v at %+v", cl.Callee().FullName(), &cl.location)
	} else {
		_, _ = fmt.Fprintf(f, "call %v at %v", cl.Callee().FullName(), &cl.location)
	}
}

func (cl *callLocation) String() string {
	return fmt.Sprintf("%v", cl)
}

func (cl *callLocation) Callee() Location {
	return cl.callee
}

func GetCallLocation(skip int) CallLocation {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return nil
	}

	callLoc := getFuncPtrLocation(pc)

	var calleeLoc Location
	var calleePc uintptr
	calleePc, _, _, ok = runtime.Caller(skip)
	if ok {
		calleeLoc = getFuncPtrLocation(calleePc)
	}

	return &callLocation{
		location: *callLoc,
		callee:   calleeLoc,
	}
}

func GetFuncLocation(f interface{}) Location {
	val := reflect.ValueOf(f)
	if val.Kind() != reflect.Func {
		return nil
	}

	return getFuncPtrLocation(val.Pointer())
}

func getFuncPtrLocation(funcPtr uintptr) *location {
	outside := runtime.FuncForPC(funcPtr)
	fileName, fileLine := outside.FileLine(funcPtr)
	fullName := outside.Name()
	lastSlash := strings.LastIndexByte(fullName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDotAfterLastSlash := strings.IndexByte(fullName[lastSlash:], '.') + lastSlash

	pkgName := fullName[:firstDotAfterLastSlash]
	funcName := fullName[firstDotAfterLastSlash+1:]

	return &location{pkgName, funcName, fileName, fileLine}
}
