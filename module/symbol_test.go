package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolEquality(t *testing.T) {
	s1 := NewSymbol("symbol1")
	s2 := NewSymbol("symbol2")

	assert.False(t, s1 == s2)
}

func TestSymbolWithSameName(t *testing.T) {
	s1 := NewSymbol("symbol")
	s2 := NewSymbol("symbol")

	assert.False(t, s1 == s2)
}

func TestSymbolInClosure(t *testing.T) {

}

func TestSymbolToString(t *testing.T) {
	s1 := NewSymbol("symbol1")

	assert.Equal(t, s1.String(), "github.com/jison/uni/provider.symbol1")
}
