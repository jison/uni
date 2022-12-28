package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jison/uni/core/valuer"

	"github.com/jison/uni/internal/errors"
	"github.com/stretchr/testify/assert"
)

func newConsumerForTest() Consumer {
	return &valueConsumer{}
}

func Test_dependency_attributes(t *testing.T) {
	t.Run("attributes", func(t *testing.T) {
		consumer := newConsumerForTest()
		d := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(1),
		}
		assert.Equal(t, consumer, d.Consumer())
		assert.False(t, d.Optional())
		assert.False(t, d.IsCollector())
		assert.Equal(t, reflect.TypeOf(1), d.Type())
		assert.Equal(t, "", d.Name())
		assert.Equal(t, 0, d.Tags().Len())
	})

	t.Run("collector type", func(t *testing.T) {
		consumer := newConsumerForTest()
		d := &dependency{
			consumer:    consumer,
			rType:       reflect.TypeOf(1),
			isCollector: true,
		}
		assert.Equal(t, reflect.TypeOf(1), d.Type())

		d2 := &dependency{
			consumer:    consumer,
			rType:       reflect.TypeOf([]int{}),
			isCollector: true,
		}
		assert.Equal(t, reflect.TypeOf(1), d2.Type())

		d3 := &dependency{
			consumer:    consumer,
			rType:       reflect.TypeOf([2]int{}),
			isCollector: true,
		}
		assert.Equal(t, reflect.TypeOf([2]int{}), d3.Type())
	})

	t.Run("Equal", func(t *testing.T) {
		tag1 := NewSymbol("tag1")

		consumer := newConsumerForTest()
		d := &dependency{
			consumer:    consumer,
			rType:       reflect.TypeOf(1),
			isCollector: true,
			optional:    true,
			name:        "abc",
			tags:        newSymbolSet(tag1),
			val:         valuer.Identity(),
		}

		t.Run("equal", func(t *testing.T) {
			d2 := d.clone()
			assert.True(t, d2.Equal(d))
		})

		t.Run("not equal to non dependency", func(t *testing.T) {
			assert.False(t, d.Equal(123))
		})

		t.Run("type", func(t *testing.T) {
			d2 := d.clone()
			d2.rType = reflect.TypeOf("")
			assert.False(t, d2.Equal(d))
		})

		t.Run("isCollector", func(t *testing.T) {
			d2 := d.clone()
			d2.isCollector = false
			assert.False(t, d2.Equal(d))
		})

		t.Run("optional", func(t *testing.T) {
			d2 := d.clone()
			d2.optional = false
			assert.False(t, d2.Equal(d))
		})

		t.Run("name", func(t *testing.T) {
			d2 := d.clone()
			d2.name = "def"
			assert.False(t, d2.Equal(d))
		})

		t.Run("tags", func(t *testing.T) {
			tag2 := NewSymbol("tag2")

			d2 := d.clone()
			d2.tags.Add(tag2)
			assert.False(t, d2.Equal(d))
		})

		t.Run("valuer", func(t *testing.T) {
			t.Run("not nil", func(t *testing.T) {
				d2 := d.clone()
				d2.val = valuer.Index(1)
				assert.False(t, d2.Equal(d))
			})

			t.Run("not nil", func(t *testing.T) {
				d2 := d.clone()
				d2.val = nil
				assert.False(t, d2.Equal(d))
				assert.False(t, d.Equal(d2))

				d3 := d.clone()
				d3.val = nil
				assert.True(t, d3.Equal(d2))
			})
		})
	})
}

func Test_dependency_builder(t *testing.T) {
	t.Run("SetOptional", func(t *testing.T) {
		d := &dependency{}
		d.SetOptional(true)
		assert.True(t, d.Dependency().Optional())

		d.SetOptional(false)
		assert.False(t, d.Dependency().Optional())
	})

	t.Run("SetAsCollector", func(t *testing.T) {
		d := &dependency{}
		d.SetAsCollector(true)
		assert.True(t, d.Dependency().IsCollector())

		d.SetAsCollector(false)
		assert.False(t, d.Dependency().IsCollector())
	})

	t.Run("SetName", func(t *testing.T) {
		d := &dependency{}
		d.SetName("abc")
		assert.Equal(t, "abc", d.Dependency().Name())

		d.SetName("def")
		assert.Equal(t, "def", d.Dependency().Name())
	})

	t.Run("AddTags", func(t *testing.T) {
		d := &dependency{}

		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")

		d.AddTags(tag1)
		assert.True(t, d.Dependency().Tags().Has(tag1))

		d.AddTags(tag2)
		assert.True(t, d.Dependency().Tags().Has(tag2))
	})
}

func Test_dependency_options(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		d := &dependency{}
		Optional(true).ApplyDependency(d)
		assert.True(t, d.Optional())

		Optional(false).ApplyDependency(d)
		assert.False(t, d.Optional())
	})

	t.Run("AsCollector", func(t *testing.T) {
		d := &dependency{}
		AsCollector(true).ApplyDependency(d)
		assert.True(t, d.IsCollector())

		AsCollector(false).ApplyDependency(d)
		assert.False(t, d.IsCollector())
	})

	t.Run("Name", func(t *testing.T) {
		d := &dependency{}
		ByName("name1").ApplyDependency(d)
		assert.Equal(t, "name1", d.Name())

		ByName("name2").ApplyDependency(d)
		assert.Equal(t, "name2", d.Name())
	})

	t.Run("Tags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")
		tag3 := NewSymbol("tag3")

		d := &dependency{}
		ByTags(tag1, tag2).ApplyDependency(d)
		assert.Equal(t, 2, d.Tags().Len())
		assert.True(t, d.Tags().Has(tag1))
		assert.True(t, d.Tags().Has(tag2))

		ByTags(tag3).ApplyDependency(d)
		assert.Equal(t, 3, d.Tags().Len())
		assert.True(t, d.Tags().Has(tag3))
	})
}

func Test_dependency_Validate(t *testing.T) {
	t.Run("not error", func(t *testing.T) {
		consumer := newConsumerForTest()
		d := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(1),
		}
		err := d.Validate()
		assert.Nil(t, err)
	})

	t.Run("with nil type", func(t *testing.T) {
		consumer := newConsumerForTest()
		d := &dependency{
			consumer: consumer,
			rType:    TypeOf(nil),
		}
		err := d.Validate()
		assert.NotNil(t, err)
	})

	t.Run("with error type", func(t *testing.T) {
		consumer := newConsumerForTest()
		d := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(errors.Newf("this is error")),
		}
		err := d.Validate()
		assert.NotNil(t, err)
	})

	t.Run("as collector while with no slice type", func(t *testing.T) {
		consumer := newConsumerForTest()
		d := &dependency{
			consumer:    consumer,
			rType:       reflect.TypeOf([2]int{}),
			isCollector: true,
		}
		err := d.Validate()
		assert.NotNil(t, err)
	})

}

func Test_dependency_clone(t *testing.T) {
	consumer := newConsumerForTest()
	tag1 := NewSymbol("tag1")
	val1 := valuer.Const(reflect.ValueOf(123))
	d1 := &dependency{
		consumer:    consumer,
		rType:       TypeOf(0),
		optional:    false,
		name:        "abc",
		tags:        newSymbolSet(tag1),
		isCollector: false,
		val:         val1,
	}

	verifyDependency := func(t *testing.T, dep Dependency) {
		assert.Equal(t, consumer, dep.Consumer())
		assert.Equal(t, TypeOf(0), dep.Type())
		assert.Equal(t, false, dep.Optional())
		assert.Equal(t, "abc", dep.Name())
		assert.Equal(t, newSymbolSet(tag1), dep.Tags())
		assert.Equal(t, false, dep.IsCollector())
		assert.Equal(t, val1, dep.Valuer())
	}

	t.Run("equality", func(t *testing.T) {
		d2 := d1.clone()
		verifyDependency(t, d2)
		assert.False(t, d2.Valuer() == d1.Valuer())
	})

	t.Run("update isolation", func(t *testing.T) {
		consumer2 := newConsumerForTest()
		tag2 := NewSymbol("tag2")
		val2 := valuer.Const(reflect.ValueOf(456))

		d2 := d1.clone()

		d2.consumer = consumer2
		d2.rType = reflect.TypeOf("")
		d2.optional = true
		d2.name = "def"
		d2.tags.Add(tag2)
		d2.isCollector = true
		d2.val = val2

		verifyDependency(t, d1)
	})

	t.Run("update isolation 2", func(t *testing.T) {
		consumer2 := newConsumerForTest()
		tag2 := NewSymbol("tag2")
		val2 := valuer.Const(reflect.ValueOf(456))

		d2 := d1.clone()

		d1.consumer = consumer2
		d1.rType = reflect.TypeOf("")
		d1.optional = true
		d1.name = "def"
		d1.tags.Add(tag2)
		d1.isCollector = true
		d1.val = val2

		verifyDependency(t, d2)
	})

	t.Run("nil", func(t *testing.T) {
		var d2 *dependency
		assert.Nil(t, d2.clone())
	})
}

func Test_dependency_Iterate(t *testing.T) {
	t.Run("continue", func(t *testing.T) {
		consumer1 := newConsumerForTest()
		tag1 := NewSymbol("tag1")
		d1 := &dependency{
			consumer:    consumer1,
			rType:       reflect.TypeOf(0),
			optional:    false,
			name:        "abc",
			tags:        newSymbolSet(tag1),
			isCollector: false,
		}

		isContinue := d1.Iterate(func(d Dependency) bool {
			assert.Equal(t, d1, d)
			return true
		})

		assert.True(t, isContinue)
	})

	t.Run("not continue", func(t *testing.T) {
		consumer1 := newConsumerForTest()
		tag1 := NewSymbol("tag1")
		d1 := &dependency{
			consumer:    consumer1,
			rType:       reflect.TypeOf(0),
			optional:    false,
			name:        "abc",
			tags:        newSymbolSet(tag1),
			isCollector: false,
		}

		isContinue := d1.Iterate(func(d Dependency) bool {
			assert.Equal(t, d1, d)
			return false
		})

		assert.False(t, isContinue)
	})
}

func Test_dependency_Format(t *testing.T) {
	t.Run("type", func(t *testing.T) {
		consumer := newConsumerForTest()
		dep := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(0),
		}
		s := fmt.Sprintf("%v", dep)
		assert.Equal(t, "Dependency[int]", s)
		vs := fmt.Sprintf("%+v", dep)
		assert.Equal(t, "Dependency[int]", vs)
	})

	t.Run("name", func(t *testing.T) {
		consumer := newConsumerForTest()
		dep := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(0),
		}
		dep.SetName("abc")
		s := fmt.Sprintf("%v", dep)
		assert.Equal(t, "Dependency[int]{name=\"abc\"}", s)
		vs := fmt.Sprintf("%+v", dep)
		assert.Equal(t, "Dependency[int]{name=\"abc\"}", vs)
	})

	t.Run("tags", func(t *testing.T) {
		consumer := newConsumerForTest()
		tag1 := NewSymbol("tag1")
		dep := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(0),
		}
		dep.AddTags(tag1)
		s := fmt.Sprintf("%v", dep)
		assert.Equal(t, "Dependency[int]{tags={tag1}}", s)
		vs := fmt.Sprintf("%+v", dep)
		vsExpected := fmt.Sprintf("Dependency[int]{tags={%+v}}", tag1)
		assert.Equal(t, vsExpected, vs)
	})

	t.Run("optional", func(t *testing.T) {
		consumer := newConsumerForTest()
		dep := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(0),
		}
		dep.SetOptional(true)
		s := fmt.Sprintf("%v", dep)
		assert.Equal(t, "Dependency[int]{optional}", s)
		vs := fmt.Sprintf("%+v", dep)
		assert.Equal(t, "Dependency[int]{optional}", vs)
	})

	t.Run("asCollector", func(t *testing.T) {
		consumer := newConsumerForTest()
		dep := &dependency{
			consumer: consumer,
			rType:    reflect.TypeOf(0),
		}
		dep.SetAsCollector(true)
		s := fmt.Sprintf("%v", dep)
		assert.Equal(t, "Dependency[int]{asCollector}", s)
		vs := fmt.Sprintf("%+v", dep)
		assert.Equal(t, "Dependency[int]{asCollector}", vs)
	})

	t.Run("String", func(t *testing.T) {
		consumer := newConsumerForTest()
		tag1 := NewSymbol("tag1")
		dep := &dependency{
			consumer:    consumer,
			rType:       reflect.TypeOf(0),
			optional:    true,
			name:        "abc",
			tags:        newSymbolSet(tag1),
			isCollector: true,
		}

		s := fmt.Sprintf("%v", dep)
		assert.Equal(t, "Dependency[int]{name=\"abc\", tags={tag1}, optional, asCollector}", s)
	})
}
