package model

import (
	"fmt"
	"testing"

	"github.com/jison/uni/core/valuer"
	"github.com/jison/uni/internal/location"
	"github.com/stretchr/testify/assert"
)

func TestLoadCriteriaConsumer(t *testing.T) {
	tag1 := NewSymbol("tag1")
	baseLoc := location.GetCallLocation(0)
	lc := LoadCriteriaConsumer(
		NewCriteria(TypeOf(1), ByName("c1"), ByTags(tag1)),
		NewCriteria(TypeOf(""), ByName("c2")),
		NewCriteria(TypeOf([]int{})).SetName("c3"),
		nil,
	)

	con := lc.Consumer()
	err := con.Validate()
	assert.Nil(t, err)

	deps := dependencyIteratorToArray(con.Dependencies())
	assert.Equal(t, 3, len(deps))
	for _, dep := range deps {
		if dep.Name() == "c1" {
			assert.Equal(t, TypeOf(1), dep.Type())
			assert.Equal(t, "c1", dep.Name())
			assert.Equal(t, newSymbolSet(tag1), dep.Tags())
			assert.Equal(t, valuer.Identity(), dep.Valuer())
			assert.False(t, dep.Optional())
			assert.False(t, dep.IsCollector())
		} else if dep.Name() == "c2" {
			assert.Equal(t, TypeOf(""), dep.Type())
			assert.Equal(t, "c2", dep.Name())
			assert.Equal(t, (*symbolSet)(nil), dep.Tags())
			assert.Equal(t, valuer.Identity(), dep.Valuer())
			assert.False(t, dep.Optional())
			assert.False(t, dep.IsCollector())
		} else if dep.Name() == "c3" {
			assert.Equal(t, TypeOf([]int{}), dep.Type())
			assert.Equal(t, "c3", dep.Name())
			assert.Equal(t, (*symbolSet)(nil), dep.Tags())
			assert.Equal(t, valuer.Identity(), dep.Valuer())
			assert.False(t, dep.Optional())
			assert.False(t, dep.IsCollector())
		}

		assert.Same(t, con, dep.Consumer())
	}

	assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), con.Valuer())
	assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
	assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
}

func Test_loadCriteriaConsumer_Consumer(t *testing.T) {
	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	lc := loadCriteriaConsumerOf(
		NewCriteria(TypeOf(1), ByName("c1"), ByTags(tag1)),
		NewCriteria(TypeOf(""), ByName("c2")),
		NewCriteria(TypeOf([]int{})).SetName("c3"),
	)
	lc.SetLocation(loc1)

	t.Run("Dependencies", func(t *testing.T) {
		con := lc.clone().Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c2" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "c2", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c3" {
				assert.Equal(t, TypeOf([]int{}), dep.Type())
				assert.Equal(t, "c3", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}
	})

	t.Run("Valuer", func(t *testing.T) {
		con := lc.clone().Consumer()
		assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), con.Valuer())
	})

	t.Run("Scope", func(t *testing.T) {
		t.Run("nil scope", func(t *testing.T) {
			con := lc.clone().Consumer()
			assert.Equal(t, GlobalScope, con.Scope())
		})

		t.Run("scope", func(t *testing.T) {
			scope1 := NewScope("scope1")
			con := lc.clone().SetScope(scope1).Consumer()
			assert.Equal(t, scope1, con.Scope())
		})
	})

	t.Run("Location", func(t *testing.T) {
		con := lc.clone().Consumer()
		assert.Equal(t, loc1, con.Location())
	})

	t.Run("Validate", func(t *testing.T) {
		t.Run("no errors", func(t *testing.T) {
			con := lc.clone().Consumer()
			err := con.Validate()
			assert.Nil(t, err)
		})

		t.Run("criteria with error type", func(t *testing.T) {
			lc2 := LoadCriteriaConsumer(NewCriteria(TypeOf((*error)(nil))))
			err := lc2.Consumer().Validate()
			assert.NotNil(t, err)
		})
	})

	t.Run("Format", func(t *testing.T) {
		con := lc.clone().Consumer()

		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("Load %v", lc.dependencies)
			assert.Equal(t, expected, fmt.Sprintf("%v", con))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("Load %v at %+v", lc.dependencies, lc.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", con))
		})
	})
}

func Test_loadCriteriaConsumer_LoadCriteriaConsumerBuilder(t *testing.T) {
	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	lc := loadCriteriaConsumerOf()

	t.Run("AddCriteria", func(t *testing.T) {
		lc1 := lc.clone()
		lc1.AddCriteria(NewCriteria(TypeOf(1), ByName("c1"), ByTags(tag1)))
		lc1.AddCriteria(NewCriteria(TypeOf(""), ByName("c2")))
		lc1.AddCriteria(NewCriteria(TypeOf([]int{})).SetName("c3"))

		con := lc1.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c2" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "c2", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c3" {
				assert.Equal(t, TypeOf([]int{}), dep.Type())
				assert.Equal(t, "c3", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}
	})

	t.Run("SetScope", func(t *testing.T) {
		scope1 := NewScope("scope1")
		lc1 := lc.clone()
		lc1.SetScope(scope1)
		con := lc1.Consumer()
		assert.Equal(t, scope1, con.Scope())
	})

	t.Run("SetLocation", func(t *testing.T) {
		lc1 := lc.clone()
		lc1.SetLocation(loc1)
		con := lc1.Consumer()
		assert.Equal(t, loc1, con.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			lc1 := lc.clone()
			lc1.SetLocation(loc1)
			lc1.UpdateCallLocation(nil)
			assert.Equal(t, loc1, lc1.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			lc1 := lc.clone()
			baseLoc := location.GetCallLocation(0)
			func() {
				lc1.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), lc1.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, lc1.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			loc := location.GetCallLocation(0)
			lc1 := lc.clone()
			lc1.UpdateCallLocation(loc)
			assert.Equal(t, loc, lc1.Location())
		})
	})

	t.Run("Consumer", func(t *testing.T) {
		lc1 := lc.clone()
		lc1.AddCriteria(NewCriteria(TypeOf(1), ByName("c1"), ByTags(tag1)))
		lc1.AddCriteria(NewCriteria(TypeOf(""), ByName("c2")))
		lc1.AddCriteria(NewCriteria(TypeOf([]int{})).SetName("c3"))
		lc1.SetLocation(loc1)

		con := lc1.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c2" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "c2", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c3" {
				assert.Equal(t, TypeOf([]int{}), dep.Type())
				assert.Equal(t, "c3", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}
		assert.Equal(t, loc1, con.Location())
		assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), con.Valuer())
	})
}

func Test_loadCriteriaConsumer_clone(t *testing.T) {
	tag1 := NewSymbol("tag1")
	loc1 := location.GetCallLocation(0)
	scope1 := NewScope("scope1")
	lc := loadCriteriaConsumerOf()
	lc.AddCriteria(NewCriteria(TypeOf(1), ByName("c1"), ByTags(tag1)))
	lc.AddCriteria(NewCriteria(TypeOf(""), ByName("c2")))
	lc.AddCriteria(NewCriteria(TypeOf([]int{})).SetName("c3"))
	lc.SetScope(scope1)
	lc.SetLocation(loc1)

	verifyConsumer := func(t *testing.T, con Consumer) {
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 3, len(deps))
		for _, dep := range deps {
			if dep.Name() == "c1" {
				assert.Equal(t, TypeOf(1), dep.Type())
				assert.Equal(t, "c1", dep.Name())
				assert.Equal(t, newSymbolSet(tag1), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c2" {
				assert.Equal(t, TypeOf(""), dep.Type())
				assert.Equal(t, "c2", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			} else if dep.Name() == "c3" {
				assert.Equal(t, TypeOf([]int{}), dep.Type())
				assert.Equal(t, "c3", dep.Name())
				assert.Equal(t, (*symbolSet)(nil), dep.Tags())
				assert.Equal(t, valuer.Identity(), dep.Valuer())
				assert.False(t, dep.Optional())
				assert.False(t, dep.IsCollector())
			}

			assert.Same(t, con, dep.Consumer())
		}

		assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), con.Valuer())
		assert.Equal(t, scope1, con.Scope())
		assert.Equal(t, loc1, con.Location())
	}

	t.Run("equality", func(t *testing.T) {
		lc2 := lc.clone()
		verifyConsumer(t, lc2.Consumer())
		assert.False(t, lc2.Valuer() == lc.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)
		scope2 := NewScope("scope2")
		lc2 := lc.clone()
		lc2.AddCriteria(NewCriteria(TypeOf([]string{})))
		lc2.SetLocation(loc2)
		lc2.SetScope(scope2)

		verifyConsumer(t, lc.Consumer())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)
		scope2 := NewScope("scope2")
		lc2 := lc.clone()
		lc3 := lc2.clone()

		lc2.AddCriteria(NewCriteria(TypeOf([]string{})))
		lc2.SetLocation(loc2)
		lc2.SetScope(scope2)

		verifyConsumer(t, lc3.Consumer())
	})
}

func TestLoadAllConsumer(t *testing.T) {
	baseLoc := location.GetCallLocation(0)
	con := LoadAllConsumer(GlobalScope).Consumer()

	err := con.Validate()
	assert.Nil(t, err)

	deps := dependencyIteratorToArray(con.Dependencies())
	assert.Equal(t, 1, len(deps))

	dep := deps[0]
	assert.Equal(t, wildcardType, dep.Type())
	assert.Equal(t, "", dep.Name())
	assert.Equal(t, (*symbolSet)(nil), dep.Tags())
	assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), dep.Valuer())
	assert.False(t, dep.Optional())
	assert.False(t, dep.IsCollector())
	assert.Same(t, con, dep.Consumer())

	assert.Equal(t, valuer.Identity(), con.Valuer())
	assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
	assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
}

func Test_loadAllConsumer_Consumer(t *testing.T) {
	scope1 := NewScope("scope1")
	baseLoc := location.GetCallLocation(0)
	con := LoadAllConsumer(scope1).Consumer()

	t.Run("Dependencies", func(t *testing.T) {
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, wildcardType, dep.Type())
		assert.Equal(t, "", dep.Name())
		assert.Equal(t, (*symbolSet)(nil), dep.Tags())
		assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), dep.Valuer())
		assert.False(t, dep.Optional())
		assert.False(t, dep.IsCollector())
		assert.Same(t, con, dep.Consumer())
	})

	t.Run("Valuer", func(t *testing.T) {
		assert.Equal(t, valuer.Identity(), con.Valuer())
	})

	t.Run("Scope", func(t *testing.T) {
		assert.Equal(t, scope1, con.Scope())
	})

	t.Run("Location", func(t *testing.T) {
		assert.Equal(t, baseLoc.FileName(), con.Location().FileName())
		assert.Equal(t, baseLoc.FileLine()+1, con.Location().FileLine())
	})

	t.Run("Validate", func(t *testing.T) {
		err := con.Validate()
		assert.Nil(t, err)
	})

	t.Run("Format", func(t *testing.T) {
		t.Run("not verbose", func(t *testing.T) {
			expected := fmt.Sprintf("LoadAll")
			assert.Equal(t, expected, fmt.Sprintf("%v", con))
		})

		t.Run("verbose", func(t *testing.T) {
			expected := fmt.Sprintf("LoadAll at %+v", con.Location())
			assert.Equal(t, expected, fmt.Sprintf("%+v", con))
		})
	})
}

func Test_loadAllConsumer_LoadAllConsumerBuilder(t *testing.T) {
	loc1 := location.GetCallLocation(0)
	lc := loadAllConsumerOf(nil)

	t.Run("SetScope", func(t *testing.T) {
		scope1 := NewScope("scope1")
		lc1 := lc.clone()
		lc1.SetScope(scope1)
		con := lc1.Consumer()
		assert.Equal(t, scope1, con.Scope())
	})

	t.Run("SetLocation", func(t *testing.T) {
		lc1 := lc.clone()
		lc1.SetLocation(loc1)
		con := lc1.Consumer()
		assert.Equal(t, loc1, con.Location())
	})

	t.Run("UpdateCallLocation", func(t *testing.T) {
		t.Run("location have been set", func(t *testing.T) {
			lc1 := lc.clone()
			lc1.SetLocation(loc1)
			lc1.UpdateCallLocation(nil)
			assert.Equal(t, loc1, lc1.Location())
		})

		t.Run("location have not been set", func(t *testing.T) {
			lc1 := lc.clone()
			baseLoc := location.GetCallLocation(0)
			func() {
				lc1.UpdateCallLocation(nil)
			}()
			assert.Equal(t, baseLoc.FileName(), lc1.Location().FileName())
			assert.Equal(t, baseLoc.FileLine()+3, lc1.Location().FileLine())
		})

		t.Run("location is not nil", func(t *testing.T) {
			loc := location.GetCallLocation(0)
			lc1 := lc.clone()
			lc1.UpdateCallLocation(loc)
			assert.Equal(t, loc, lc1.Location())
		})
	})

	t.Run("Consumer", func(t *testing.T) {
		scope1 := NewScope("scope1")
		lc1 := lc.clone()
		lc1.SetLocation(loc1)
		lc1.SetScope(scope1)

		con := lc1.Consumer()

		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, wildcardType, dep.Type())
		assert.Equal(t, "", dep.Name())
		assert.Equal(t, (*symbolSet)(nil), dep.Tags())
		assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), dep.Valuer())
		assert.False(t, dep.Optional())
		assert.False(t, dep.IsCollector())
		assert.Same(t, con, dep.Consumer())

		assert.Equal(t, valuer.Identity(), con.Valuer())
		assert.Equal(t, loc1, con.Location())
		assert.Equal(t, scope1, con.Scope())
		assert.Equal(t, valuer.Identity(), con.Valuer())
	})
}

func Test_loadAllConsumer_clone(t *testing.T) {
	loc1 := location.GetCallLocation(0)
	scope1 := NewScope("scope1")
	lc := loadAllConsumerOf(nil)
	lc.SetScope(scope1)
	lc.SetLocation(loc1)

	verifyConsumer := func(t *testing.T, con Consumer) {
		deps := dependencyIteratorToArray(con.Dependencies())
		assert.Equal(t, 1, len(deps))

		dep := deps[0]
		assert.Equal(t, wildcardType, dep.Type())
		assert.Equal(t, "", dep.Name())
		assert.Equal(t, (*symbolSet)(nil), dep.Tags())
		assert.Equal(t, valuer.Collector(TypeOf((*interface{})(nil))), dep.Valuer())
		assert.False(t, dep.Optional())
		assert.False(t, dep.IsCollector())
		assert.Same(t, con, dep.Consumer())

		assert.Equal(t, valuer.Identity(), con.Valuer())
		assert.Equal(t, loc1, con.Location())
		assert.Equal(t, scope1, con.Scope())
		assert.Equal(t, valuer.Identity(), con.Valuer())
	}

	t.Run("equality", func(t *testing.T) {
		lc2 := lc.clone()
		verifyConsumer(t, lc2.Consumer())
		assert.False(t, lc2.Valuer() == lc.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)
		scope2 := NewScope("scope2")
		lc2 := lc.clone()
		lc2.SetLocation(loc2)
		lc2.SetScope(scope2)

		verifyConsumer(t, lc.Consumer())
	})

	t.Run("update isolation 2", func(t *testing.T) {
		loc2 := location.GetCallLocation(0)
		scope2 := NewScope("scope2")
		lc2 := lc.clone()
		lc3 := lc2.clone()

		lc2.SetLocation(loc2)
		lc2.SetScope(scope2)

		verifyConsumer(t, lc3.Consumer())
	})
}
