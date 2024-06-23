package structmorph

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

const tmpl = `// Code generated by structmorph; DO NOT EDIT.

package {{.ToPkg}}

{{if ne .FromPkg "main"}}import "{{.FromPkgPath}}"{{end}}

func {{.FuncNameToDTO}}(src {{if ne .FromPkg "main"}}{{.FromPkg}}.{{end}}{{.From}}) {{.To}} {
	return {{.To}}{
		{{range .Fields}}{{.ToField}}: src.{{.FromField}},
		{{end}}
	}
}

func {{.FuncNameToStruct}}(src {{.To}}) {{if ne .FromPkg "main"}}{{.FromPkg}}.{{end}}{{.From}} {
	return {{if ne .FromPkg "main"}}{{.FromPkg}}.{{end}}{{.From}}{
		{{range .Fields}}{{.FromField}}: src.{{.ToField}},
		{{end}}
	}
}
`

func Generate(from, to string) error {
	fromStructName, err := ParseStructName(from)
	if err != nil {
		return err
	}
	toStructName, err := ParseStructName(to)
	if err != nil {
		return err
	}

	fromStruct, err := FindAndParseStruct(fromStructName)
	if err != nil {
		return err
	}
	slog.Info("Found and parsed struct", slog.Any("struct", fromStruct))

	toStruct, err := FindAndParseStruct(toStructName)
	if err != nil {
		return err
	}
	slog.Info("Found and parsed struct", slog.Any("struct", toStruct))

	fields, err := CreateMapping(fromStruct, toStruct)
	if err != nil {
		return err
	}

	data := TemplateData{
		FuncNameToDTO:    fmt.Sprintf("ConvertTo%s", toStructName.Name),
		FuncNameToStruct: fmt.Sprintf("ConvertTo%s", fromStructName.Name),
		FromPkg:          fromStruct.Package,
		FromPkgPath:      fromStruct.ImportPath,
		From:             fromStruct.Name,
		To:               toStruct.Name,
		ToPkg:            toStruct.Package,
		Fields:           fields,
	}

	buff := &bytes.Buffer{}
	err = data.GenerateCode(buff)
	if err != nil {
		return fmt.Errorf("error generating code: %w", err)
	}

	fileName := fromStruct.FileName()
	err = FormatAndWrite(buff, fileName)
	if err != nil {
		return fmt.Errorf("error formatting and writing code: %w", err)
	}

	slog.Info("Generated and formatted code", "file", fileName)
	return nil
}

type StructName struct {
	Package string
	Name    string
}

func ParseStructName(rawName string) (StructName, error) {
	rawName = strings.TrimSpace(rawName)
	if rawName == "" {
		return StructName{}, fmt.Errorf("empty input")
	}

	parts := strings.Split(rawName, ".")
	if len(parts) == 1 {
		return StructName{
			Package: "main",
			Name:    parts[0],
		}, nil
	}

	if len(parts) != 2 {
		return StructName{}, fmt.Errorf("invalid format for --from or --to. Expected 'package.StructName'")
	}
	return StructName{
		Package: parts[0],
		Name:    parts[1],
	}, nil
}

type StructType struct {
	StructName
	ImportPath string
	Fields     map[string]string
}

func FindAndParseStruct(name StructName) (s StructType, err error) {
	cfg := &packages.Config{
		//todo убрать потом то что не нужно
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedCompiledGoFiles | packages.NeedDeps | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return StructType{}, fmt.Errorf("error loading packages: %w", err)
	}

	var found bool
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.GenDecl:
					for _, spec := range node.Specs {
						if t, ok := spec.(*ast.TypeSpec); ok && t.Name.Name == name.Name {
							filePath := pkg.Fset.Position(file.Pos()).Filename
							s = parseStructType(name, pkg, t)
							found = true
							slog.Info("Struct found in file", "struct", s.Name, "file", filePath, "importPath", pkg.PkgPath, "fields", s.Fields)
							return false
						}
					}
				}
				return true
			})
		}
	}

	if !found {
		return s, fmt.Errorf("struct not found")
	}

	return
}

func parseStructType(name StructName, pkg *packages.Package, spec *ast.TypeSpec) StructType {
	s := StructType{
		StructName: name,
	}
	s.ImportPath = pkg.PkgPath
	s.Fields = extractFields(spec)

	return s
}

func extractFields(t *ast.TypeSpec) map[string]string {
	fields := map[string]string{}
	for _, field := range t.Type.(*ast.StructType).Fields.List {
		fieldName := field.Names[0].Name
		fieldType := fmt.Sprintf("%s", field.Type)
		fields[fieldName] = fieldType
	}
	return fields
}

type FieldMapping struct {
	FromField string
	ToField   string
}

func CreateMapping(fromStruct, toStruct StructType) ([]FieldMapping, error) {
	var fields []FieldMapping
	for fieldName, fieldType := range toStruct.Fields {
		fromFieldType, ok := fromStruct.Fields[fieldName]
		if !ok || fromFieldType != fieldType {
			return nil, fmt.Errorf("field not found or type mismatch, field: %s, type: %s, struct: %s", fieldName, fieldType, fromStruct.Name)
		}
		fields = append(fields, FieldMapping{
			FromField: fieldName,
			ToField:   fieldName,
		})
	}

	return fields, nil
}

func FormatAndWrite(buff *bytes.Buffer, fileName string) error {
	// Run goimports on the generated file
	formattedSource, err := imports.Process(fileName, buff.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("error running goimports on generated file: %w", err)
	}

	// Write the formatted source back to the file
	err = os.WriteFile(fileName, formattedSource, 0644)
	if err != nil {
		return fmt.Errorf("error writing formatted source to file: %w", err)
	}

	return nil
}

type TemplateData struct {
	FuncNameToDTO    string
	FuncNameToStruct string
	FromPkg          string
	FromPkgPath      string
	From             string
	ToPkg            string
	To               string
	Fields           []FieldMapping
}

func (data TemplateData) GenerateCode(output io.Writer) error {
	t := template.Must(template.New("morph").Parse(tmpl))
	err := t.Execute(output, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func (s StructType) FileName() string {
	return fmt.Sprintf("morph_%s.go", strings.ToLower(s.Name))
}
