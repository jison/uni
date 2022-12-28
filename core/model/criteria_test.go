package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCriteria(t *testing.T) {
	type structForCriteriaTest struct {
		_ int
	}

	t.Run("NewCriteria", func(t *testing.T) {
		cri := NewCriteria(reflect.TypeOf(structForCriteriaTest{})).Criteria()
		assert.NotNil(t, cri)
		assert.Equal(t, cri.Type(), reflect.TypeOf(structForCriteriaTest{}))
		assert.Equal(t, cri.Name(), "")
		assert.Equal(t, 0, cri.Tags().Len())
	})

	t.Run("add name option", func(t *testing.T) {
		cri := NewCriteria(reflect.TypeOf(structForCriteriaTest{}), ByName("name1")).Criteria()
		assert.NotNil(t, cri)
		assert.Equal(t, "name1", cri.Name())
	})

	t.Run("override name", func(t *testing.T) {
		cri := NewCriteria(
			reflect.TypeOf(structForCriteriaTest{}), ByName("name1"), ByName("name2"),
		).Criteria()
		assert.NotNil(t, cri)
		assert.Equal(t, "name2", cri.Name())
	})

	t.Run("add tags option", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		cri := NewCriteria(reflect.TypeOf(structForCriteriaTest{}), ByTags(tag1)).Criteria()
		assert.NotNil(t, cri)
		assert.Equal(t, 1, cri.Tags().Len())
		assert.True(t, cri.Tags().Has(tag1))
	})

	t.Run("add multiple tags option", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")
		tag3 := NewSymbol("tag3")
		cri := NewCriteria(
			reflect.TypeOf(structForCriteriaTest{}), ByTags(tag1, tag2), ByTags(tag3),
		).Criteria()
		assert.NotNil(t, cri)
		assert.Equal(t, 3, cri.Tags().Len())
		assert.True(t, cri.Tags().Has(tag1))
		assert.True(t, cri.Tags().Has(tag2))
		assert.True(t, cri.Tags().Has(tag3))
	})

	t.Run("set name", func(t *testing.T) {
		cri := NewCriteria(structForCriteriaTest{}).SetName("abc").Criteria()
		assert.Equal(t, "abc", cri.Name())
	})

	t.Run("add tags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")
		cri := NewCriteria(structForCriteriaTest{}).AddTags(tag1, tag2).Criteria()
		assert.True(t, cri.Tags().Has(tag1))
		assert.True(t, cri.Tags().Has(tag2))
	})
}

func Test_criteria_Format(t *testing.T) {
	t.Run("without name and tags", func(t *testing.T) {
		cri := &criteria{
			rType: reflect.TypeOf(1),
			name:  "",
		}

		assert.Equal(t, "{type=int}", fmt.Sprintf("%v", cri))
		assert.Equal(t, "{type=int}", fmt.Sprintf("%+v", cri))
	})

	t.Run("with name and without tags", func(t *testing.T) {
		cri := &criteria{
			rType: reflect.TypeOf(1),
			name:  "name",
		}

		assert.Equal(t, "{type=int, name=\"name\"}", fmt.Sprintf("%v", cri))
		assert.Equal(t, "{type=int, name=\"name\"}", fmt.Sprintf("%+v", cri))
	})

	t.Run("without name and with tags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		cri := &criteria{
			rType: reflect.TypeOf(1),
			tags:  newSymbolSet(tag1),
			name:  "",
		}

		assert.Equal(t, "{type=int, tags={tag1}}", fmt.Sprintf("%v", cri))
		assert.Equal(t, "{type=int, tags={github.com/jison/uni/core/model.Test_criteria_Format.func3.tag1}}",
			fmt.Sprintf("%+v", cri))
	})

	t.Run("with name and tags", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		cri := &criteria{
			rType: reflect.TypeOf(1),
			tags:  newSymbolSet(tag1),
			name:  "name",
		}

		assert.Equal(t, "{type=int, name=\"name\", tags={tag1}}", fmt.Sprintf("%v", cri))
		assert.Equal(t, "{type=int, name=\"name\", tags={"+
			"github.com/jison/uni/core/model.Test_criteria_Format.func4.tag1}}", fmt.Sprintf("%+v", cri))
	})

}

func Test_criteria_clone(t *testing.T) {
	t.Run("equality", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")
		cri := &criteria{
			rType: reflect.TypeOf(1),
			tags:  newSymbolSet(tag1, tag2),
			name:  "abc",
		}

		cri2 := cri.clone()

		assert.Equal(t, cri.Type(), cri2.Type())
		assert.Equal(t, cri.Tags(), cri2.Tags())
		assert.Equal(t, cri.Name(), cri2.Name())
	})

	t.Run("update isolation", func(t *testing.T) {
		tag1 := NewSymbol("tag1")
		tag2 := NewSymbol("tag2")

		cri := &criteria{
			rType: reflect.TypeOf(1),
			tags:  newSymbolSet(tag1),
			name:  "abc",
		}

		cri2 := cri.clone()
		cri2.AddTags(tag2)
		cri2.SetName("def")

		assert.Equal(t, cri.Type(), cri2.Type())
		assert.Equal(t, newSymbolSet(tag1), cri.Tags())
		assert.Equal(t, "abc", cri.Name())
	})

	t.Run("nil", func(t *testing.T) {
		var c2 *criteria
		assert.Nil(t, c2.clone())
	})
}

func Test_criteria_equal(t *testing.T) {
	tag1 := NewSymbol("tag1")
	tag2 := NewSymbol("tag2")

	c1 := &criteria{
		rType: reflect.TypeOf(1),
		tags:  newSymbolSet(tag1),
		name:  "abc",
	}

	t.Run("equal", func(t *testing.T) {
		c2 := c1.clone()
		assert.True(t, c2.Equal(c1))
	})

	t.Run("not equal to non criteria", func(t *testing.T) {
		assert.False(t, c1.Equal(123))
	})

	t.Run("nil equal nil", func(t *testing.T) {
		var c2 *criteria
		var c3 *criteria
		assert.True(t, c2.Equal(c3))
	})

	t.Run("type", func(t *testing.T) {
		c2 := c1.clone()
		c2.rType = reflect.TypeOf("")
		assert.False(t, c2.Equal(c1))
	})

	t.Run("name", func(t *testing.T) {
		c2 := c1.clone()
		c2.name = "def"
		assert.False(t, c2.Equal(c1))
	})

	t.Run("tags", func(t *testing.T) {
		c2 := c1.clone()
		c2.tags.Add(tag2)
		assert.False(t, c2.Equal(c1))
	})
}
