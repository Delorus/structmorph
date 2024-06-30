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

func (p *Parser) FindAndParseStructDst(name StructName) (DstStructType, error) {
	result := &DstStructType{StructName: name}
	return *result, p.FindStruct(name, func(name StructName, pkg *packages.Package, spec *ast.TypeSpec) {
		result.FilePath = filepath.Dir(pkg.Fset.Position(spec.Pos()).Filename)
		result.extractFields(spec)
	})
}

func (p *Parser) FindAndParseStructSrc(name StructName) (SrcStructType, error) {
	result := &SrcStructType{StructName: name}
	return *result, p.FindStruct(name, func(name StructName, pkg *packages.Package, spec *ast.TypeSpec) {
		result.ImportPath = pkg.Types.Path()
		result.extractFields(spec)
	})
}

func (t *SrcStructType) extractFields(spec *ast.TypeSpec) {
	list := spec.Type.(*ast.StructType).Fields.List
	fields := make(map[string]SrcFieldType, len(list))
	for _, field := range list {
		fieldType := parseFieldType(field.Type)

		fieldName := fieldType.Name
		// handle anonymous struct fields
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		}

		fields[fieldName] = SrcFieldType{
			FieldType{
				Name: fieldName,
				Type: fieldType,
			},
		}
	}
	t.Fields = fields
}

func (s *DstStructType) extractFields(t *ast.TypeSpec) {
	list := t.Type.(*ast.StructType).Fields.List
	fields := make([]DstFieldType, 0, len(list))
	for _, astField := range list {
		fieldType := parseFieldType(astField.Type)

		fieldName := fieldType.Name
		// handle anonymous struct fields
		if len(astField.Names) > 0 {
			fieldName = astField.Names[0].Name
		}

		field := DstFieldType{
			FieldType: FieldType{
				Name: fieldName,
				Type: fieldType,
			},
			SrcField: fieldName,
		}

		if astField.Tag != nil {
			tag := astField.Tag.Value
			if strings.HasPrefix(tag, "`morph:") {
				tagValue := strings.Trim(tag, "`")
				tagValue = strings.TrimPrefix(tagValue, "morph:\"")
				tagValue = strings.TrimSuffix(tagValue, "\"")
				field.SrcField = tagValue
			}
		}

		//}
		//for _, tag := range astField.Names[0].Obj.Decl.(*ast.Field).Tag.Value {
		// find tag starts with `morph:"`
		fields = append(fields, field)
	}

	s.Fields = fields
}

func parseFieldType(field ast.Expr) FieldTypeType {
	switch t := field.(type) {
	case *ast.Ident:
		return FieldTypeType{Name: t.Name}
	case *ast.ArrayType:
		return FieldTypeType{Name: fmt.Sprintf("[]%s", parseFieldType(t.Elt).Name)} //todo limit recursion
	case *ast.MapType:
		return FieldTypeType{Name: fmt.Sprintf("map[%s]%s", parseFieldType(t.Key).Name, parseFieldType(t.Value).Name)}
	case *ast.StarExpr:
		return FieldTypeType{Name: parseFieldType(t.X).Name, IsPointer: true} //todo test how it would works with two pointers
	//case *ast.SelectorExpr:
	//	return fmt.Sprintf("%s.%s", t.X, t.Sel)
	default:
		return FieldTypeType{Name: fmt.Sprintf("%T", t)}
	}
}
