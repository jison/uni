package model

import "fmt"

type DependencyIterator interface {
	Iterate(func(Dependency) bool) bool
}

type emptyDependencyIterator struct{}

func (e emptyDependencyIterator) Iterate(_ func(Dependency) bool) bool { return true }
func (e emptyDependencyIterator) Format(s fmt.State, r rune)           { formatDependencyIterator(e, s, r) }

type ArrayDependencyIterator []Dependency

func (i ArrayDependencyIterator) Iterate(f func(Dependency) bool) bool {
	for _, d := range i {
		if !f(d) {
			return false
		}
	}
	return true
}

func (i ArrayDependencyIterator) Format(s fmt.State, r rune) { formatDependencyIterator(i, s, r) }

type combinedDependencyIterator struct {
	left  DependencyIterator
	right DependencyIterator
}

func (i *combinedDependencyIterator) Iterate(f func(Dependency) bool) bool {
	if !i.left.Iterate(f) {
		return false
	}
	return i.right.Iterate(f)
}

func (i *combinedDependencyIterator) Format(s fmt.State, r rune) { formatDependencyIterator(i, s, r) }

func CombineDependencyIterators(its ...DependencyIterator) DependencyIterator {
	if len(its) == 0 {
		return emptyDependencyIterator{}
	}

	lastIt := its[0]
	for _, it := range its[1:] {
		lastIt = &combinedDependencyIterator{
			left:  lastIt,
			right: it,
		}
	}

	return lastIt
}

func formatDependencyIterator(it DependencyIterator, s fmt.State, r rune) {
	_, _ = fmt.Fprint(s, "[")
	var depFormat string
	if s.Flag('+') && r == 'v' {
		depFormat = "%+v"
	} else {
		depFormat = "%v"
	}

	first := true
	it.Iterate(func(dep Dependency) bool {
		if !first {
			_, _ = fmt.Fprint(s, ", ")
		} else {
			first = false
		}

		_, _ = fmt.Fprintf(s, depFormat, dep)
		return true
	})
	_, _ = fmt.Fprint(s, "]")
}
