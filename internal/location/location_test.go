package location

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// baselineFunc can use this function to get the start line of code
func baselineFunc() {}

func ExportedFunc()   {}
func UnexportedFunc() {}
func getClosures() (func(), func(), func()) {
	closure1 := func() {}
	closure2 := func() func() {
		return func() {}
	}()
	closure3 := func() func() func() {
		return func() func() {
			return func() {}
		}
	}()()
	return closure1, closure2, closure3
}

func GetCallLocationFromExportedFunc() CallLocation {
	return GetCallLocation(1)
}

func getCallLocationFromUnexportedFunc() CallLocation {
	return GetCallLocation(1)
}

func getClosuresOfGettingCallLocation() (func() CallLocation, func() CallLocation, func() CallLocation) {
	closure1 := func() CallLocation {
		return GetCallLocation(1)
	}
	closure2 := func() func() CallLocation {
		return func() CallLocation {
			return GetCallLocation(1)
		}
	}()
	closure3 := func() func() func() CallLocation {
		return func() func() CallLocation {
			return func() CallLocation {
				return GetCallLocation(1)
			}
		}
	}()()

	return closure1, closure2, closure3
}

func getCallLocationInDeepCall() CallLocation {
	return func() CallLocation {
		return func() CallLocation {
			return GetCallLocation(3)
		}()
	}()
}

func TestGetFuncLocation(t *testing.T) {
	type want struct {
		pkgName    string
		funcName   string
		fileSuffix string
		fileLine   int
	}

	closure1, closure2, closure3 := getClosures()

	baseLoc := GetFuncLocation(baselineFunc)

	tests := []struct {
		name string
		fn   func()
		want want
	}{
		{"exported function", ExportedFunc, want{
			"github.com/jison/uni/internal/location",
			"ExportedFunc",
			"internal/location/location_test.go",
			baseLoc.FileLine() + 2,
		}},
		{"unexported function", UnexportedFunc, want{
			"github.com/jison/uni/internal/location",
			"UnexportedFunc",
			"internal/location/location_test.go",
			baseLoc.FileLine() + 3,
		}},
		{"closure1", closure1, want{
			"github.com/jison/uni/internal/location",
			"getClosures.func1",
			"internal/location/location_test.go",
			baseLoc.FileLine() + 5,
		}},
		{"closure2", closure2, want{
			"github.com/jison/uni/internal/location",
			"getClosures.func4",
			"internal/location/location_test.go",
			baseLoc.FileLine() + 7,
		}},
		{"closure3", closure3, want{
			"github.com/jison/uni/internal/location",
			"getClosures.func7",
			"internal/location/location_test.go",
			baseLoc.FileLine() + 11,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := GetFuncLocation(tt.fn)
			assert.Equal(t, tt.want.pkgName, l.PkgName())
			assert.Equal(t, tt.want.funcName, l.FuncName())
			assert.True(t, strings.HasSuffix(l.FileName(), tt.want.fileSuffix))
			assert.Equal(t, tt.want.fileLine, l.FileLine())
		})
	}

	t.Run("function is nil", func(t *testing.T) {
		l := GetFuncLocation(nil)
		assert.Nil(t, l)
	})
}

func TestLocation_Format(t *testing.T) {
	l := &location{
		pkgName:  "this/is/package",
		funcName: "i_am_function",
		fileName: "file_name.go",
		fileLine: 7,
	}

	assert.Equal(t, "file_name.go:7", l.String())
	assert.Equal(t, "file_name.go:7", fmt.Sprintf("%v", l))
	assert.Equal(t, "file_name.go:7", fmt.Sprintf("%s", l))

	assert.Equal(t, "this/is/package.i_am_function (file_name.go:7)", fmt.Sprintf("%+v", l))
}

func TestGetCallerLocation(t *testing.T) {
	type want struct {
		pkgName  string
		funcName string
	}

	closure1, closure2, closure3 := getClosuresOfGettingCallLocation()

	tests := []struct {
		name string
		fn   func() CallLocation
		want want
	}{
		{"exported function", GetCallLocationFromExportedFunc, want{
			"github.com/jison/uni/internal/location",
			"GetCallLocationFromExportedFunc",
		}},
		{"unexported function", getCallLocationFromUnexportedFunc, want{
			"github.com/jison/uni/internal/location",
			"getCallLocationFromUnexportedFunc",
		}},
		{"closure1", closure1, want{
			"github.com/jison/uni/internal/location",
			"getClosuresOfGettingCallLocation.func1",
		}},
		{"closure2", closure2, want{
			"github.com/jison/uni/internal/location",
			"getClosuresOfGettingCallLocation.func4",
		}},
		{"closure3", closure3, want{
			"github.com/jison/uni/internal/location",
			"getClosuresOfGettingCallLocation.func3.1.1",
		}},
		{"get location in deep call", getCallLocationInDeepCall, want{
			"github.com/jison/uni/internal/location",
			"getCallLocationInDeepCall",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			baseLoc := GetFuncLocation(func() {})

			l := tt.fn()
			assert.Equal(t, "github.com/jison/uni/internal/location", l.PkgName())
			assert.Equal(t, "TestGetCallerLocation.func1", l.FuncName())
			assert.True(t, strings.HasSuffix(l.FileName(), "internal/location/location_test.go"))
			assert.Equal(t, baseLoc.FileLine()+2, l.FileLine())
			assert.Equal(t, tt.want.pkgName, l.Callee().PkgName())
			assert.Equal(t, tt.want.funcName, l.Callee().FuncName())
		})
	}

	t.Run("can not get call location", func(t *testing.T) {
		loc := GetCallLocation(1000000)
		assert.Nil(t, loc)
	})
}

func TestCallLocation_Format(t *testing.T) {
	cl := &callLocation{
		location: location{
			pkgName:  "this/is/package",
			funcName: "i_am_function",
			fileName: "fileName.go",
			fileLine: 7,
		},
		callee: &location{
			pkgName:  "callee/package",
			funcName: "callee_function",
			fileName: "callee_fileName.go",
			fileLine: 11,
		},
	}

	assert.Equal(t, "call callee/package.callee_function at fileName.go:7", cl.String())
	assert.Equal(t, "call callee/package.callee_function at fileName.go:7", fmt.Sprintf("%v", cl))
	assert.Equal(t, "call callee/package.callee_function at fileName.go:7", fmt.Sprintf("%s", cl))

	assert.Equal(t, "call callee/package.callee_function at this/is/package.i_am_function (fileName.go:7)",
		fmt.Sprintf("%+v", cl))
}
