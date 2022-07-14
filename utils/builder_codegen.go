package utils

import (
	"bytes"
	"fmt"
	"go/format"
	"reflect"
	"text/template"

	"strings"
	"unicode"

	"github.com/zoumo/goset"
)

func GenBuilderForType(targetType reflect.Type) (string, error) {
	fieldList := getTypeFieldsOfGetter(targetType)
	if len(fieldList) == 0 {
		return "", fmt.Errorf("type `%s` has no getter method", targetType.Name())
	}

	imports := goset.NewSetFromStrings([]string{})
	for _, field := range fieldList {
		path := field.Type.PkgPath()
		if path != "" {
			imports.Add(path)
		}
	}

	info := typeInfo{
		Name:    targetType.Name(),
		PkgName: pkgNameOfPkgPath(targetType.PkgPath()),
		PkgPath: targetType.PkgPath(),
		Imports: imports.ToStrings(),
		Fields:  fieldList,
	}
	return genBuilderSourceCode(info)
}

func pkgNameOfPkgPath(pkgPath string) string {
	slashIdx := strings.LastIndex(pkgPath, "/")
	if slashIdx == -1 {
		return pkgPath
	} else {
		return pkgPath[slashIdx+1:]
	}
}

func getTypeFieldsOfGetter(t reflect.Type) []*field {
	var fList []*field
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if method.Type.NumIn() != 0 || method.Type.NumOut() != 1 || method.Type.IsVariadic() {
			continue
		}
		f := &field{
			Name: method.Name,
			Type: method.Type.Out(0),
		}
		fList = append(fList, f)
	}
	return fList
}

var getterTpl = `// DO NOT EDIT!!!
package {{.PkgName}}

import (
{{range $name := .Imports}}
	"{{$name}}"
{{end}}
)

{{$structName := (printf "%sStruct" .Name)}}

{{$builderInterfaceName := (printf "%sBuilder" .Name)}}

type {{$builderInterfaceName}} interface {
	{{.Name}}
{{range .Fields}}
	With{{.Name}}(value {{.Type}}) {{$builderInterfaceName}}
{{- end}}
	Result() {{.Name}}
}

type {{$structName}} struct {
{{range .Fields}}
	{{.Name}}Value {{.Type}}
{{- end}}
}

{{range .Fields}}
func (v *{{$structName}}) {{.Name}}() {{.Type}} {
	return v.{{.Name}}Value
}

func (v *{{$structName}}) With{{.Name}}(value {{.Type}}) *{{$structName}} {
	v.{{.Name}}Value = value
	return v
}
{{end}}

func (v *{{$structName}}) Result() {{.Name}} {
	return v
}
`

func genBuilderSourceCode(info typeInfo) (string, error) {

	tpl := template.New("builder").Funcs(template.FuncMap{
		"Capitalize": func(s string) string {
			return strings.Title(s)
		},
		"Uncapitalize": func(s string) string {
			return string(unicode.ToLower(rune(s[0]))) + s[1:]
		},
	})
	tpl = template.Must((tpl.Parse(getterTpl)))
	var buf bytes.Buffer
	tpl.Execute(&buf, info)
	// fmt.Printf("%s", buf.String())
	formatedBuf, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("format template: %w", err)
	}
	return string(formatedBuf), nil
}

type typeInfo struct {
	Name    string
	PkgName string
	PkgPath string
	Imports []string
	Fields  []*field
}

type field struct {
	Name string
	Type reflect.Type
}
