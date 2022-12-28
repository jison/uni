package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := Newf("abc")
	assert.Equal(t, err.Error(), "abc")

	err2 := New(err)
	assert.Equal(t, err2.Error(), "abc")
}

func TestBugError(t *testing.T) {
	err := Bug(Newf("abc"))
	assert.Equal(t, "looks like you have found a bug in uni.\n\tabc", err.Error())

	err2 := Bugf("abc")
	assert.Equal(t, "looks like you have found a bug in uni.\n\tabc", err2.Error())
}

func Test_structError_HasError(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		var e *structError
		assert.False(t, e.HasError())
	})

	t.Run("have main error", func(t *testing.T) {
		e := Empty().WithMainf("main error")
		assert.True(t, e.HasError())
	})

	t.Run("have sub error", func(t *testing.T) {
		e := Empty().AddErrorf("sub error")
		assert.True(t, e.HasError())
	})

	t.Run("no error", func(t *testing.T) {
		e := Empty()
		assert.False(t, e.HasError())

		e2 := &structError{}
		assert.False(t, e2.HasError())
	})
}

func Test_structError_WithMain(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		var e0 *structError
		e1 := Newf("main error")
		e2 := e0.WithMain(e1)
		assert.ErrorIs(t, e2, e1)
	})

	t.Run("main is nil", func(t *testing.T) {
		var e0 *structError
		e1 := e0.WithMain(nil)
		assert.False(t, e1.HasError())
	})

	t.Run("single line", func(t *testing.T) {
		se1 := Empty()
		se2 := se1.WithMain(Newf("abc"))
		se3 := se2.WithMain(Newf("efg"))
		assert.Nil(t, se1)
		assert.False(t, se1.HasError())
		assert.True(t, se2.HasError())
		assert.Equal(t, "abc", se2.Error())
		assert.True(t, se3.HasError())
		assert.Equal(t, "efg", se3.Error())

		se4 := se3.WithMain(nil)
		assert.Nil(t, se4)
		assert.False(t, se4.HasError())
	})

	t.Run("multiple lines", func(t *testing.T) {
		se1 := Empty()
		se2 := se1.WithMainf("abc\ndef\nght")
		assert.Equal(t, "abc\ndef\nght", se2.Error())
	})
}

func Test_structError_AddSubError(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		se1 := Empty()
		se2 := se1.AddErrors(Newf("abc"))
		se3 := se2.AddErrors(Newf("efg"))
		se4 := se3.AddErrors(nil, nil)
		assert.Nil(t, se1)
		assert.False(t, se1.HasError())
		assert.True(t, se2.HasError())
		assert.Equal(t, "abc", se2.Error())
		assert.True(t, se3.HasError())
		assert.Equal(t, "[0] abc\n[1] efg", se3.Error())
		assert.Equal(t, "[0] abc\n[1] efg", se4.Error())

		se11 := se1.AddErrors(nil)
		assert.Nil(t, se11)
		assert.False(t, se11.HasError())
	})

	t.Run("multiple lines", func(t *testing.T) {
		se1 := Empty()
		se2 := se1.AddErrors(Newf("abc\ndef"))
		se3 := se2.AddErrors(Newf("efg\nhij"))

		assert.Equal(t, "abc\ndef", se2.Error())
		assert.Equal(t, "[0] abc\n    def\n[1] efg\n    hij", se3.Error())
	})
}

func Test_structError_WithMainAndSub(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		se1 := Empty()
		se2 := se1.WithMainf("abc")
		se3 := se2.AddErrorf("def")
		se4 := se3.AddErrorf("ghi")

		assert.Equal(t, "abc\n\t[0] def\n\t[1] ghi", se4.Error())
	})

	t.Run("simple case 2", func(t *testing.T) {
		se1 := Empty().WithMainf("se1")
		se2 := Empty().WithMainf("se2").AddErrors(se1)
		se3 := Empty().WithMainf("se3").AddErrors(se2)
		se4 := Empty().WithMainf("se4").AddErrors(se3)

		expected := "se4\n" +
			"\tse3\n" +
			"\t\tse2\n" +
			"\t\t\tse1"

		assert.Equal(t, expected, se4.Error())
	})

	t.Run("complicate case", func(t *testing.T) {
		e111 := Empty().AddErrorf("e111.S1").AddErrorf("e111.S2")
		e121 := Empty().WithMainf("e121.M").AddErrorf("e121.2")
		e12 := Empty().WithMainf("e12.M").AddErrorf("e12.S1").
			AddErrorf("e12.S2\ne12.S2L2\ne12.S2L3").
			AddErrors(e121)
		e11 := Empty().AddErrors(e111)
		e1 := Empty().AddErrors(e11).AddErrors(e12)

		expected := "[0] e111.S1\n" +
			"[1] e111.S2\n" +
			"[2] e12.M\n" +
			"\t[2.0] e12.S1\n" +
			"\t[2.1] e12.S2\n" +
			"\t      e12.S2L2\n" +
			"\t      e12.S2L3\n" +
			"\t[2.2] e121.M\n" +
			"\t\t[2.2.0] e121.2"
		assert.Equal(t, expected, e1.Error())
	})
}

func Test_structError_As(t *testing.T) {
	t.Run("as *structError", func(t *testing.T) {
		se1 := &structError{
			mainError: Newf("abc"),
			subErrors: []error{Newf("def"), Newf("ghi")},
		}
		var se2 *structError
		ok := se1.As(&se2)
		assert.True(t, ok)
		assert.Same(t, se1, se2)
	})

	t.Run("as StructError", func(t *testing.T) {
		se1 := &structError{
			mainError: Newf("abc"),
			subErrors: []error{Newf("def"), Newf("ghi")},
		}
		var se2 StructError
		ok := se1.As(&se2)
		assert.True(t, ok)
		assert.Same(t, se1, se2)
	})

	t.Run("main", func(t *testing.T) {
		e := &testError{msg: "abc"}

		se1 := &structError{
			mainError: e,
		}

		var e2 *testError
		ok := se1.As(&e2)
		assert.True(t, ok)
		assert.Same(t, e, e2)
	})

	t.Run("sub", func(t *testing.T) {
		e := &testError{msg: "abc"}

		se1 := &structError{
			mainError: Newf("abc"),
			subErrors: []error{Newf("def"), e},
		}

		var e2 *testError
		ok := se1.As(&e2)
		assert.True(t, ok)
		assert.Same(t, e, e2)
	})
}

func TestIs(t *testing.T) {
	t.Run("empty is empty", func(t *testing.T) {
		se1 := Empty()
		se2 := Empty()
		assert.True(t, Is(se1, se2))
	})

	t.Run("main error", func(t *testing.T) {
		e := Newf("abc")
		se1 := Empty().WithMain(e)
		se2 := Empty().WithMain(e)
		assert.True(t, Is(se1, se2))
	})

	t.Run("sub error", func(t *testing.T) {
		e := Newf("abc")
		se1 := Empty().AddErrors(e)
		se2 := Empty().AddErrors(e)
		assert.True(t, Is(se1, se2))
	})

	t.Run("complicated case", func(t *testing.T) {
		e1 := Newf("abc")
		e2 := Newf("def")
		se1 := Empty().AddErrors(e1)
		se2 := Empty().AddErrors(se1)
		se3 := Empty().AddErrors(e1, e2)
		se4 := Empty().AddErrors(se3)

		assert.True(t, Is(se2, se4))
		assert.True(t, Is(se4, se2))
		assert.True(t, Is(se4, e2))
		assert.False(t, Is(se2, e2))
	})
}

func TestAs(t *testing.T) {
	t.Run("as *structError", func(t *testing.T) {
		se1 := &structError{
			mainError: Newf("abc"),
			subErrors: []error{Newf("def"), Newf("ghi")},
		}
		var se2 *structError
		ok := As(se1, &se2)
		assert.True(t, ok)
		assert.Same(t, se1, se2)
	})

	t.Run("as StructError", func(t *testing.T) {
		se1 := &structError{
			mainError: Newf("abc"),
			subErrors: []error{Newf("def"), Newf("ghi")},
		}
		var se2 StructError
		ok := As(se1, &se2)
		assert.True(t, ok)
		assert.Same(t, se1, se2)
	})

	t.Run("main", func(t *testing.T) {
		e := &testError{msg: "abc"}

		se1 := &structError{
			mainError: e,
		}

		var e2 *testError
		ok := As(se1, &e2)
		assert.True(t, ok)
		assert.Same(t, e, e2)
	})

	t.Run("sub", func(t *testing.T) {
		e := &testError{msg: "abc"}

		se1 := &structError{
			mainError: Newf("abc"),
			subErrors: []error{Newf("def"), e},
		}

		var e2 *testError
		ok := As(se1, &e2)
		assert.True(t, ok)
		assert.Same(t, e, e2)
	})
}

type testError struct {
	msg string
}

func (t *testError) Error() string {
	return t.msg
}

func TestStructError_Unwrap(t *testing.T) {
	t.Run("have main", func(t *testing.T) {
		e := Newf("abc")
		e2 := &structError{
			mainError: e,
		}
		assert.Same(t, e, e2.Unwrap())
	})

	t.Run("have sub", func(t *testing.T) {
		e := Newf("abc")
		e2 := &structError{
			subErrors: []error{e},
		}
		assert.Same(t, e, e2.Unwrap())
	})

	t.Run("empty", func(t *testing.T) {
		e := &structError{}
		assert.Nil(t, e.Unwrap())
	})

	t.Run("nil", func(t *testing.T) {
		var e *structError
		assert.Nil(t, e.Unwrap())
	})
}
