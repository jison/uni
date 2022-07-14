package module

import (
	"runtime"
	"strings"
)

// Symbol unique identifier for Tags
type Symbol interface {
	String() string
	Id() interface{}
}

type symbol struct {
	name     string
	pkg      string
	value    *symbol
	fileName string
	fileLine int
}

func (s *symbol) String() string {
	return s.pkg + "." + s.name
}

func (s *symbol) Id() interface{} {
	return s
}

func NewSymbol(name string) Symbol {
	t := symbol{}
	t.name = name
	t.value = &t

	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		// panic("can not get `NewTag` caller")
		return &t
	}

	outside := runtime.FuncForPC(pc)
	t.fileName, t.fileLine = outside.FileLine(pc)
	funcName := outside.Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash

	pkgName := funcName[:lastDot]
	t.pkg = pkgName

	return t.value
}
