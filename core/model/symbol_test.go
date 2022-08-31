package model

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbol(t *testing.T) {
	t.Run("equality", func(t *testing.T) {
		s1 := NewSymbol("symbol1")
		s2 := NewSymbol("symbol2")

		assert.False(t, s1 == s2)
	})

	t.Run("symbols with same name", func(t *testing.T) {
		s1 := NewSymbol("symbol")
		s2 := NewSymbol("symbol")

		assert.False(t, s1 == s2)
	})

	t.Run("string", func(t *testing.T) {
		s1 := NewSymbol("symbol1")

		assert.Equal(t, "github.com/jison/uni/core/model.TestSymbol.func3.symbol1", fmt.Sprintf("%+v", s1))
		assert.Equal(t, "symbol1", fmt.Sprintf("%v", s1))
	})
}

func Test_symbolSet_Has(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set *symbolSet
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			assert.False(t, set.Has(s))
		}
	})

	t.Run("before and after adding symbol", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			assert.False(t, set.Has(s))
			set.Add(s)
			assert.True(t, set.Has(s))
		}
	})

	t.Run("before and after deleting symbol", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		set.Iterate(func(s Symbol) bool {
			assert.True(t, set.Has(s))
			set.Del(s)
			assert.False(t, set.Has(s))
			return true
		})
	})
}

func Test_symbolSet_Len(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set *symbolSet
		assert.Equal(t, 0, set.Len())
	})

	t.Run("before and after adding symbol", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			assert.Equal(t, i, set.Len())
			set.Add(s)
			assert.Equal(t, i+1, set.Len())
		}
	})

	t.Run("add symbol already in set", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		set.Iterate(func(s Symbol) bool {
			l := set.Len()
			set.Add(s)
			assert.Equal(t, l, set.Len())
			return true
		})
	})

	t.Run("before add after deleting symbol", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		set.Iterate(func(s Symbol) bool {
			l := set.Len()
			set.Del(s)
			assert.Equal(t, l-1, set.Len())
			return true
		})

		assert.Equal(t, 0, set.Len())
	})
}

func Test_symbolSet_Iterate(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set *symbolSet
		r := set.Iterate(func(s Symbol) bool {
			t.Fail()
			return true
		})
		assert.True(t, r)
	})

	t.Run("iterate all", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		symbols := make(map[Symbol]struct{})
		set.Iterate(func(s Symbol) bool {
			symbols[s] = struct{}{}
			return true
		})

		assert.Equal(t, 10, len(symbols))
	})

	t.Run("interrupt iteration", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		symbols := make(map[Symbol]struct{})
		set.Iterate(func(s Symbol) bool {
			symbols[s] = struct{}{}

			return len(symbols) < 5
		})

		assert.Equal(t, 5, len(symbols))
	})
}

func Test_symbolSet_Equal(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set1 *symbolSet
		var set2 *symbolSet
		assert.True(t, set1.Equal(set2))
	})

	t.Run("equal", func(t *testing.T) {
		symbols := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			symbols.Add(s)
		}

		set1 := newSymbolSet()
		set2 := newSymbolSet()

		symbols.Iterate(func(s Symbol) bool {
			set1.Add(s)
			set2.Add(s)
			assert.True(t, set1.Equal(set2))

			return true
		})

		symbols.Iterate(func(s Symbol) bool {
			set1.Del(s)
			set2.Del(s)
			assert.True(t, set1.Equal(set2))

			return true
		})
	})

	t.Run("not equal", func(t *testing.T) {
		set1 := newSymbolSet()
		set2 := newSymbolSet()

		s1 := NewSymbol("s1")
		s2 := NewSymbol("s2")

		set1.Add(s1)
		set2.Add(s2)
		assert.False(t, set1.Equal(set2))
	})

	t.Run("nil equal empty", func(t *testing.T) {
		var set1 *symbolSet
		set2 := newSymbolSet()
		assert.True(t, set1.Equal(set2))
	})
}

func Test_symbolSet_Add(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set *symbolSet
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			assert.False(t, set.Has(s))
			set.Add(s)
			assert.False(t, set.Has(s))
		}
	})

	t.Run("add symbols", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
			assert.True(t, set.Has(s))
			assert.Equal(t, i+1, set.Len())
		}
	})

	t.Run("add symbol that already in set", func(t *testing.T) {
		symbols := []Symbol{
			NewSymbol("s1"), NewSymbol("s2"), NewSymbol("s3"),
		}

		set := newSymbolSet(symbols...)

		set.Add(symbols[1])
		assert.Equal(t, symbols, set.symbols())
	})

	t.Run("add nil symbol", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}
		oldSymbols := set.symbols()

		set.Add(nil)
		assert.False(t, set.Has(nil))
		assert.Equal(t, 10, set.Len())
		assert.Equal(t, oldSymbols, set.symbols())
	})
}

func Test_symbolSet_Del(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set *symbolSet
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			assert.False(t, set.Has(s))
			set.Del(s)
			assert.False(t, set.Has(s))
		}
	})

	t.Run("delete symbols", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 10; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		set.Iterate(func(s Symbol) bool {
			l := set.Len()
			set.Del(s)
			assert.Equal(t, l-1, set.Len())
			assert.False(t, set.Has(s))
			return true
		})
	})
}

func Test_symbolSet_Format(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var set *symbolSet
		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{}", s)
		vs := fmt.Sprintf("%+v", set)
		assert.Equal(t, "{}", vs)
	})

	t.Run("empty", func(t *testing.T) {
		set := newSymbolSet()
		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{}", s)
		vs := fmt.Sprintf("%+v", set)
		assert.Equal(t, "{}", vs)
	})

	t.Run("1 symbol", func(t *testing.T) {
		set := newSymbolSet()
		set.Add(NewSymbol("s"))

		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{s}", s)
		vs := fmt.Sprintf("%+v", set)
		assert.Equal(t, "{github.com/jison/uni/core/model.Test_symbolSet_Format.func3.s}", vs)
	})

	t.Run("n symbol", func(t *testing.T) {
		set := newSymbolSet()
		for i := 0; i < 2; i++ {
			s := NewSymbol("s" + strconv.Itoa(i))
			set.Add(s)
		}

		s := fmt.Sprintf("%v", set)
		assert.Equal(t, "{s0, s1}", s)
		vs := fmt.Sprintf("%+v", set)
		vsExpected := "{" +
			"github.com/jison/uni/core/model.Test_symbolSet_Format.func4.s0" +
			", " +
			"github.com/jison/uni/core/model.Test_symbolSet_Format.func4.s1" +
			"}"
		assert.Equal(t, vsExpected, vs)
	})
}

func Test_symbolSet_clone(t *testing.T) {
	t.Run("symbolSet is nil", func(t *testing.T) {
		var set *symbolSet
		set2 := set.clone()
		assert.Nil(t, set2)
	})

	t.Run("equality", func(t *testing.T) {
		s1 := NewSymbol("s1")
		s2 := NewSymbol("s2")
		s3 := NewSymbol("s3")
		set := newSymbolSet(s1, s2, s3)
		set2 := set.clone()

		assert.Equal(t, set.symbols(), set2.symbols())
		assert.Equal(t, set.indexBySymbol, set2.indexBySymbol)
		assert.Equal(t, set.lastIndex, set2.lastIndex)
	})

	t.Run("update isolation", func(t *testing.T) {
		s1 := NewSymbol("s1")
		s2 := NewSymbol("s2")
		s3 := NewSymbol("s3")
		s4 := NewSymbol("s4")
		set := newSymbolSet(s1, s2, s3)
		set2 := set.clone()

		set2.Add(s4)
		assert.Equal(t, []Symbol{s1, s2, s3}, set.symbols())
		assert.Equal(t, 3, set.lastIndex)
		assert.Equal(t, map[Symbol]int{s1: 0, s2: 1, s3: 2}, set.indexBySymbol)

		set3 := set.clone()
		set.Add(s4)

		assert.Equal(t, []Symbol{s1, s2, s3}, set3.symbols())
		assert.Equal(t, 3, set3.lastIndex)
		assert.Equal(t, map[Symbol]int{s1: 0, s2: 1, s3: 2}, set3.indexBySymbol)
	})
}

func Test_newSymbolSet(t *testing.T) {
	t.Run("empty symbols", func(t *testing.T) {
		set := newSymbolSet()

		assert.Equal(t, 0, set.Len())
	})

	t.Run("one symbol", func(t *testing.T) {
		s := NewSymbol("s")
		set := newSymbolSet(s)

		assert.Equal(t, 1, set.Len())
		assert.True(t, set.Has(s))
	})

	t.Run("multiple symbols", func(t *testing.T) {
		symbols := []Symbol{
			NewSymbol("s1"), NewSymbol("s2"), NewSymbol("s3"),
		}

		set := newSymbolSet(symbols...)

		assert.Equal(t, len(symbols), set.Len())
		for _, s := range symbols {
			set.Has(s)
		}
	})
}
