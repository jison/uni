package load

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"sort"

	"github.com/jison/uni/module"
	"golang.org/x/tools/go/packages"
)

type Loader interface {
	Load() (module.Module, error)
}

type ModuleSpec struct {
	PkgPath string
	Module  *packages.Module
}

var moduleIFace = reflect.TypeOf(struct{ module.Module }{}).Field(0).Type

type DefaultLoader struct {
	// Path is the path for the Iface package.
	Path string
	// Names are the type names to load. Empty means all types in the package.
	Names []string
}

func (l *DefaultLoader) Load() (module.Module, error) {
	return nil, nil
}

func (l *DefaultLoader) readModules() (*ModuleSpec, error) {
	loadCfg := packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
	}
	pkgs, err := packages.Load(&loadCfg, l.Path, moduleIFace.PkgPath())
	if err != nil {
		return nil, fmt.Errorf("loading package: %w", err)
	}
	if len(pkgs) < 2 {
		return nil, fmt.Errorf("missing package information for: %s", l.Path)
	}
	entPkg, pkg := pkgs[0], pkgs[1]
	if len(pkg.Errors) != 0 {
		return nil, pkg.Errors[0]
	}
	if pkgs[0].PkgPath != moduleIFace.PkgPath() {
		entPkg, pkg = pkgs[1], pkgs[0]
	}
	var names []string
	iface := entPkg.Types.Scope().Lookup(moduleIFace.Name()).Type().Underlying().(*types.Interface)
	for k, v := range pkg.TypesInfo.Defs {
		typ, ok := v.(*types.TypeName)
		if !ok || !k.IsExported() || !types.Implements(typ.Type(), iface) {
			continue
		}
		spec, ok := k.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil, fmt.Errorf("invalid declaration %T for %s", k.Obj.Decl, k.Name)
		}
		if _, ok := spec.Type.(*ast.StructType); !ok {
			return nil, fmt.Errorf("invalid spec type %T for %s", spec.Type, k.Name)
		}
		names = append(names, k.Name)
	}
	if len(l.Names) == 0 {
		l.Names = names
	}
	sort.Strings(l.Names)
	return &ModuleSpec{PkgPath: pkg.PkgPath, Module: pkg.Module}, nil
}
