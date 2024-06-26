package structmorph

import (
	"fmt"
	"go/ast"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

type Parser struct {
	ProjectRoot string

	pkgCache struct {
		once       sync.Once
		pkgs       []*packages.Package
		loadPkgErr error
	}
}

type ParseStructTypeFunc func(name StructName, pkg *packages.Package, spec *ast.TypeSpec)

func (p *Parser) FindStruct(name StructName, parser ParseStructTypeFunc) error {
	cfg := &packages.Config{
		Dir:  p.ProjectRoot,
		Logf: log.Printf, //todo
		//todo убрать потом то что не нужно
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedCompiledGoFiles | packages.NeedDeps | packages.NeedImports,
	}

	p.pkgCache.once.Do(func() {
		pkgs, err := packages.Load(cfg, "./...")
		if err != nil {
			p.pkgCache.loadPkgErr = fmt.Errorf("error loading packages: %w", err)
		}
		p.pkgCache.pkgs = pkgs
	})
	if p.pkgCache.loadPkgErr != nil {
		return p.pkgCache.loadPkgErr
	}

	var found bool
	for _, pkg := range p.pkgCache.pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.GenDecl:
					for _, spec := range node.Specs {
						if t, ok := spec.(*ast.TypeSpec); ok && t.Name.Name == name.Name {
							filePath := pkg.Fset.Position(file.Pos()).Filename
							parser(name, pkg, t)
							found = true
							slog.Info("Struct found in file", "struct", name, "file", filePath, "importPath", pkg.Types.Path())
							return false
						}
					}
				}
				return true
			})
		}
	}

	if !found {
		return fmt.Errorf("struct not found")
	}

	return nil
}

func (p *Parser) FindAndParseStructTo(name StructName) (ToStructType, error) {
	result := &ToStructType{StructName: name}
	return *result, p.FindStruct(name, func(name StructName, pkg *packages.Package, spec *ast.TypeSpec) {
		result.FilePath = filepath.Dir(pkg.Fset.Position(spec.Pos()).Filename)
		result.extractFields(spec)
	})
}

func (p *Parser) FindAndParseStructFrom(name StructName) (FromStructType, error) {
	result := &FromStructType{StructName: name}
	return *result, p.FindStruct(name, func(name StructName, pkg *packages.Package, spec *ast.TypeSpec) {
		result.ImportPath = pkg.Types.Path()
		result.extractFields(spec)
	})
}

func (t *FromStructType) extractFields(spec *ast.TypeSpec) {
	list := spec.Type.(*ast.StructType).Fields.List
	fields := make(map[string]string, len(list))
	for _, field := range list {
		fieldName := field.Names[0].Name
		fieldType := fmt.Sprintf("%s", field.Type)
		fields[fieldName] = fieldType
	}
	t.Fields = fields
}

func (s *ToStructType) extractFields(t *ast.TypeSpec) {
	list := t.Type.(*ast.StructType).Fields.List
	fields := make([]ToFieldType, 0, len(list))
	for _, field := range list {
		fieldName := field.Names[0].Name
		fieldType := ToFieldType{
			Name:      fieldName,
			Type:      fmt.Sprintf("%s", field.Type),
			FromField: fieldName,
		}

		if field.Tag != nil {
			tag := field.Tag.Value
			if strings.HasPrefix(tag, "`morph:") {
				tagValue := strings.Trim(tag, "`")
				tagValue = strings.TrimPrefix(tagValue, "morph:\"")
				tagValue = strings.TrimSuffix(tagValue, "\"")
				fieldType.FromField = tagValue
			}
		}

		//}
		//for _, tag := range field.Names[0].Obj.Decl.(*ast.Field).Tag.Value {
		// find tag starts with `morph:"`
		fields = append(fields, fieldType)
	}

	s.Fields = fields
}
