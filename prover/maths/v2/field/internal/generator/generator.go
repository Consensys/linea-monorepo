package main

import (
	"fmt"
	"go/types"
	"io"
	"path"
	"text/template"

	_ "embed"

	"github.com/consensys/linea-monorepo/prover/utils"
	"golang.org/x/tools/go/packages"
)

//go:embed template.go.tmpl
var templateSource string

// GeneratorConfig is the config provided to the generator
type GeneratorConfig struct {
	NewPackageName    string        `json:"newPackageName"`
	GeneratedFilename string        `json:"generatedFilename"`
	Mapping           []TypeMapping `json:"mappings"`
	ExtraImports      []string      `json:"extraImports"`
}

// TypeMapping is the input mapping of the generator. It maps a source type and
// a package to a new type name.
type TypeMapping struct {
	SrcType string `json:"srcType"`
	SrcPkg  string `json:"srcPkg"`
	NewType string `json:"newType"`
}

// TemplateValues stores the values to use to generate the template.
type TemplateValues struct {
	NewPackageName string
	Packages       []PackageInfo
	Methods        []MethodInfo
	Types          []TypeMapping
}

// PackageInfo stores the information about an imported package
type PackageInfo struct {
	Name string
	Path string
}

// MethodInfo collects all the necessary information needed to generate a copy
// method.
type MethodInfo struct {
	MethodName string
	Receiver   FuncParams
	Params     []FuncParams
	Results    []FuncParams
}

// CollectedType is the result of the collector
type CollectedType struct {
	Type    *types.Named
	Pkg     *packages.Package
	Methods []*types.Func
}

func Generate(w io.Writer, cfg *GeneratorConfig) error {

	collectedTypes, err := collectTypesAndMethods(cfg.Mapping)
	if err != nil {
		return err
	}

	values := &TemplateValues{
		NewPackageName: cfg.NewPackageName,
		Packages:       []PackageInfo{},
		Methods:        []MethodInfo{},
		Types:          []TypeMapping{},
	}

	for i := range cfg.ExtraImports {
		values.addPackage(&packages.Package{
			PkgPath: cfg.ExtraImports[i],
			Name:    path.Base(cfg.ExtraImports[i]),
		})
	}

	for i := range collectedTypes {
		values.addPackage(collectedTypes[i].Pkg)
		values.addTypeDeclaration(cfg.Mapping[i].NewType, &collectedTypes[i])
	}

	for i := range collectedTypes {
		for j := range collectedTypes[i].Methods {
			values.addMethodInfo(cfg, collectedTypes[i].Methods[j])
		}
	}

	return fillTemplate(w, values)
}

func fillTemplate(w io.Writer, values *TemplateValues) error {

	// Parse and execute template
	tmpl, err := template.New("code").Funcs(template.FuncMap{
		"renderMethod": RenderMethodDefinition,
		"pathBase":     path.Base,
	}).Parse(templateSource)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err := tmpl.Execute(w, values); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

func collectTypesAndMethods(mapping []TypeMapping) ([]CollectedType, error) {

	// Load the source package
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.LoadImports,
	}

	res := make([]CollectedType, 0, len(mapping))

	for _, mapEntry := range mapping {

		pkgs, err := packages.Load(cfg, mapEntry.SrcPkg)
		if err != nil {
			return nil, fmt.Errorf("loading package: %w", err)
		}
		if len(pkgs) != 1 {
			return nil, fmt.Errorf("expected 1 package, got %d", len(pkgs))
		}

		pkg := pkgs[0]
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package errors: %v", pkg.Errors)
		}

		fmt.Printf("collecting package with name %s and path %s\n", pkg.Name, pkg.PkgPath)

		// Find the source type
		obj := pkg.Types.Scope().Lookup(mapEntry.SrcType)
		if obj == nil {
			return nil, fmt.Errorf("type %s not found in package %s", mapEntry.SrcType, mapEntry.SrcPkg)
		}

		named, ok := obj.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf("%s is not a named type", mapEntry.SrcType)
		}

		collected := CollectedType{
			Type: named,
			Pkg:  pkg,
		}

		for i := 0; i < named.NumMethods(); i++ {
			method := named.Method(i)
			if !method.Exported() {
				continue
			}

			collected.Methods = append(collected.Methods, method)
		}

		res = append(res, collected)
	}

	return res, nil
}

// addPackage adds a package if it was not already added
func (t *TemplateValues) addPackage(pkg *packages.Package) {

	fmt.Printf("adding package %s\n", pkg.PkgPath)

	for i := range t.Packages {

		if t.Packages[i].Path == pkg.PkgPath {
			return
		}

		if t.Packages[i].Name == pkg.Name {
			utils.Panic("packages have same name but different paths: `%v` and `%v`", pkg.PkgPath, t.Packages[i].Path)
		}
	}

	t.Packages = append(t.Packages, PackageInfo{
		Name: pkg.Name,
		Path: pkg.PkgPath,
	})
}

func (tmpl *TemplateValues) addTypeDeclaration(newType string, t *CollectedType) {
	tmpl.Types = append(tmpl.Types, TypeMapping{
		SrcType: t.Type.Obj().Name(),
		SrcPkg:  t.Pkg.PkgPath,
		NewType: newType,
	})
}

// addMethodInfo builds a [MethodInfo] from a [types.Func]
func (tmpl *TemplateValues) addMethodInfo(cfg *GeneratorConfig, fn *types.Func) {

	methodInfo := MethodInfo{
		MethodName: fn.Name(),
		Receiver:   cfg.newFuncParams(fn.Signature().Recv()),
		Params:     make([]FuncParams, fn.Signature().Params().Len()),
		Results:    make([]FuncParams, fn.Signature().Results().Len()),
	}

	for i := range methodInfo.Params {
		methodInfo.Params[i] = cfg.newFuncParams(fn.Signature().Params().At(i))
	}

	for i := range methodInfo.Results {
		methodInfo.Results[i] = cfg.newFuncParams(fn.Signature().Results().At(i))
	}

	tmpl.Methods = append(tmpl.Methods, methodInfo)
}

// newFuncParams builds a [FuncParams] from a [types.Signature]
func (cfg *GeneratorConfig) newFuncParams(v *types.Var) FuncParams {

	info := FuncParams{
		Name:    v.Name(),
		SrcType: NewFromTypesPkg(v.Type()),
		NewType: NewMappedType(v.Type(), cfg.Mapping),
	}

	if info.Name == "" || info.Name == "_" {
		info.Name = "arg"
	}

	return info
}
