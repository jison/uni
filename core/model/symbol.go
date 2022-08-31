package model

import (
	"fmt"
	"sort"

	"github.com/jison/uni/internal/location"
)

// Symbol unique identifier for Tags
type Symbol interface {
	ID() interface{} // ID make sure every symbol is unique
}

type SymbolSet interface {
	Has(Symbol) bool
	Len() int
	Iterate(func(Symbol) bool) bool
	Equal(SymbolSet) bool
}

type symbol struct {
	name  string
	value *symbol
	loc   location.Location
}

func (s *symbol) ID() interface{} {
	return s
}

func (s *symbol) Format(f fmt.State, verb rune) {
	if f.Flag('+') && verb == 'v' && s.loc != nil {
		_, _ = fmt.Fprintf(f, "%s.%s", s.loc.FullName(), s.name)
	} else {
		_, _ = fmt.Fprintf(f, "%s", s.name)
	}
}

func (s *symbol) String() string {
	return fmt.Sprintf("%v", s)
}

func NewSymbol(name string) Symbol {
	t := symbol{}
	t.name = name
	t.value = &t // prevent the symbol clean by gc

	loc := location.GetCallLocation(1)
	t.loc = loc

	return t.value
}

func newSymbolSet(ss ...Symbol) *symbolSet {
	set := &symbolSet{}
	for _, s := range ss {
		set.Add(s)
	}
	return set
}

type symbolSet struct {
	indexBySymbol map[Symbol]int // record the order in which symbols are added
	lastIndex     int
}

func (ss *symbolSet) Has(s Symbol) bool {
	if ss == nil {
		return false
	}

	_, ok := ss.indexBySymbol[s]
	return ok
}

func (ss *symbolSet) Len() int {
	if ss == nil {
		return 0
	}

	return len(ss.indexBySymbol)
}

func (ss *symbolSet) Iterate(f func(Symbol) bool) bool {
	if ss == nil {
		return true
	}

	for s := range ss.indexBySymbol {
		if !f(s) {
			return false
		}
	}

	return true
}

func (ss *symbolSet) Equal(other SymbolSet) bool {
	if ss.Len() != other.Len() {
		return false
	}

	return other.Iterate(func(s Symbol) bool {
		return ss.Has(s)
	})
}

func (ss *symbolSet) Add(s Symbol) {
	if ss == nil || s == nil {
		return
	}

	if ss.indexBySymbol == nil {
		ss.indexBySymbol = map[Symbol]int{}
	}
	if _, ok := ss.indexBySymbol[s]; ok {
		return
	}
	ss.indexBySymbol[s] = ss.lastIndex
	ss.lastIndex += 1
}

func (ss *symbolSet) Del(s Symbol) {
	if ss == nil {
		return
	}

	delete(ss.indexBySymbol, s)
}

func (ss *symbolSet) symbols() []Symbol {
	if ss == nil {
		return nil
	}

	var symbols []Symbol
	for s := range ss.indexBySymbol {
		symbols = append(symbols, s)
	}

	sort.Slice(symbols, func(i, j int) bool {
		return ss.indexBySymbol[symbols[i]] < ss.indexBySymbol[symbols[j]]
	})

	return symbols
}

func (ss *symbolSet) clone() *symbolSet {
	if ss == nil {
		return nil
	}

	set := &symbolSet{
		indexBySymbol: map[Symbol]int{},
		lastIndex:     ss.lastIndex,
	}

	for s, i := range ss.indexBySymbol {
		set.indexBySymbol[s] = i
	}

	return set
}

func (ss *symbolSet) Format(fs fmt.State, r rune) {
	var symbolFormat string
	if fs.Flag('+') && r == 'v' {
		symbolFormat = "%+v"
	} else {
		symbolFormat = "%v"
	}

	_, _ = fmt.Fprint(fs, "{")
	first := true
	for _, s := range ss.symbols() {
		if !first {
			_, _ = fmt.Fprintf(fs, ", ")
		} else {
			first = false
		}

		_, _ = fmt.Fprintf(fs, symbolFormat, s)
	}
	_, _ = fmt.Fprint(fs, "}")
}
