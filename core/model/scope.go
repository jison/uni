package model

import (
	"fmt"

	"github.com/jison/uni/internal/location"
)

type Scope interface {
	// Name a word to recognize the scope
	Name() string
	// ID makes sure every scope is unique
	ID() interface{}
	// CanEnterFrom check if this scope is the descendant of another scope
	CanEnterFrom(other Scope) bool
	// CanEnterDirectlyFrom check if this scope can enter directly from another scope
	// other scope can enter directly into this scope,
	// only if the other scope is the parent of this scope
	CanEnterDirectlyFrom(other Scope) bool
}

type scope struct {
	name    string
	id      *scope
	loc     location.Location
	parents map[Scope]struct{}
}

func (s *scope) Name() string {
	return s.name
}

func (s *scope) ID() interface{} {
	return s.id
}

func (s *scope) CanEnterFrom(s2 Scope) bool {
	if s == s2 {
		return false
	}
	if len(s.parents) == 0 {
		return false
	}

	for p := range s.parents {
		if p == s2 {
			return true
		}

		r := p.CanEnterFrom(s2)
		if r {
			return true
		}
	}

	return false
}

func (s *scope) CanEnterDirectlyFrom(other Scope) bool {
	_, ok := s.parents[other]
	return ok
}

func (s *scope) Format(f fmt.State, c rune) {
	if f.Flag('+') && c == 'v' {
		_, _ = fmt.Fprintf(f, "%v.%v", s.loc.FullName(), s.Name())
	} else {
		_, _ = fmt.Fprint(f, s.Name())
	}
}

func (s *scope) String() string {
	return fmt.Sprintf("%v", s)
}

type globalScope struct{}

var GlobalScope = &globalScope{}

func (g *globalScope) Name() string {
	return "Global"
}

func (g *globalScope) ID() interface{} {
	return GlobalScope
}

func (g *globalScope) CanEnterFrom(_ Scope) bool {
	return false
}

func (g *globalScope) CanEnterDirectlyFrom(_ Scope) bool {
	return false
}

func (g *globalScope) Format(f fmt.State, _ rune) {
	_, _ = fmt.Fprint(f, g.Name())
}

func NewScope(name string, parents ...Scope) Scope {
	s := scope{
		name: name,
	}
	s.id = &s // prevent the scope clean by gc
	s.loc = location.GetCallLocation(1)
	s.parents = map[Scope]struct{}{}

	if len(parents) > 0 {
		for _, p := range parents {
			s.parents[p] = struct{}{}
		}
	} else {
		s.parents[GlobalScope] = struct{}{}
	}

	return s.id
}
