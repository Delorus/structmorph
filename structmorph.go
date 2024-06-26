package structmorph

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

type GenerationConfig struct {
	ProjectRoot string
}

func (c *GenerationConfig) NewParser() *Parser {
	return &Parser{
		ProjectRoot: c.ProjectRoot,
	}
}

func DefaultGenerationConfig() *GenerationConfig {
	return &GenerationConfig{
		ProjectRoot: ".",
	}
}

type GenerationConfigOption func(*GenerationConfig)

func WithProjectRoot(root string) GenerationConfigOption {
	return func(cfg *GenerationConfig) {
		cfg.ProjectRoot = root
	}
}

func Generate(from, to string, opts ...GenerationConfigOption) error {
	cfg := DefaultGenerationConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	parser := cfg.NewParser()

	fromStructName, err := ParseStructName(from)
	if err != nil {
		return err
	}
	toStructName, err := ParseStructName(to)
	if err != nil {
		return err
	}

	fromStruct, err := parser.FindAndParseStructFrom(fromStructName)
	if err != nil {
		return err
	}
	slog.Info("Found and parsed struct", slog.Any("struct", fromStruct))

	toStruct, err := parser.FindAndParseStructTo(toStructName)
	if err != nil {
		return err
	}
	slog.Info("Found and parsed struct", slog.Any("struct", toStruct))

	data, err := CreateTemplateData(fromStruct, toStruct)
	if err != nil {
		return fmt.Errorf("error creating template data: %w", err)
	}

	buff := &bytes.Buffer{}
	err = data.GenerateCode(buff)
	if err != nil {
		return fmt.Errorf("error generating code: %w", err)
	}

	fileName := filepath.Join(toStruct.Filepath(), fromStruct.FileName())
	err = FormatAndWrite(buff, fileName)
	if err != nil {
		return fmt.Errorf("error formatting and writing code: %w", err)
	}

	slog.Info("Generated and formatted code", "file", fileName)
	return nil
}

func CreateTemplateData(fromStruct FromStructType, toStruct ToStructType) (TemplateData, error) {
	data := TemplateData{
		FuncNameToDTO:    fmt.Sprintf("ConvertTo%s", toStruct.Name),
		FuncNameToStruct: fmt.Sprintf("ConvertTo%s", fromStruct.Name),
		FromStructName:   fromStruct.Name,
		ToStructName:     toStruct.Name,
		DistFilePkgName:  toStruct.Package,
	}

	data.FromStructName = resolveSrcStructName(fromStruct, toStruct)
	data.FromPkgPathImport = resolveSrcStructImport(fromStruct, toStruct)

	fields, err := CreateMapping(fromStruct, toStruct)
	if err != nil {
		return data, err
	}
	data.Fields = fields

	return data, nil
}

func resolveSrcStructName(fromStruct FromStructType, toStruct ToStructType) string {
	if fromStruct.Package == toStruct.Package {
		return fromStruct.Name
	}
	return fmt.Sprintf("%s.%s", fromStruct.Package, fromStruct.Name)
}

func resolveSrcStructImport(fromStruct FromStructType, toStruct ToStructType) string {
	if fromStruct.Package == toStruct.Package {
		return ""
	}
	return fromStruct.ImportPath
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

type ToStructType struct {
	StructName
	Fields   []ToFieldType
	FilePath string
}

func (s *ToStructType) Filepath() string {
	return s.FilePath
}

type FromStructType struct {
	StructName
	ImportPath string
	Fields     map[string]string
}

type ToFieldType struct {
	Name      string
	Type      string
	FromField string
}

type FieldMapping struct {
	FromField string
	ToField   ToFieldType
}

func CreateMapping(fromStruct FromStructType, toStruct ToStructType) ([]FieldMapping, error) {
	var fields []FieldMapping
	for _, toField := range toStruct.Fields {
		fromField := toField.FromField
		fromFieldType, ok := fromStruct.Fields[fromField]
		if !ok || fromFieldType != toField.Type {
			return nil, fmt.Errorf("field not found or type mismatch, field: %s, type: %+v, struct: %s", fromField, toField, fromStruct.Name)
		}
		fields = append(fields, FieldMapping{
			FromField: fromField,
			ToField:   toField,
		})
	}

	return fields, nil
}

func FormatAndWrite(buff *bytes.Buffer, fileName string) error {
	// Run goimports on the generated file
	formattedSource, err := imports.Process(fileName, buff.Bytes(), nil)
	if err != nil {
		slog.Error("Error running goimports on generated file", "error", err)
		// Write the unformatted source to the file anyway, so the user can see what went wrong
		formattedSource = buff.Bytes()
	}

	// Write the formatted source back to the file
	err = os.WriteFile(fileName, formattedSource, 0644)
	if err != nil {
		return fmt.Errorf("error writing formatted source to file: %w", err)
	}

	return nil
}

type TemplateData struct {
	FuncNameToDTO     string
	FuncNameToStruct  string
	FromPkgPathImport string
	FromStructName    string
	DistFilePkgName   string
	ToStructName      string
	Fields            []FieldMapping
}

const tmpl = `// Code generated by structmorph; DO NOT EDIT.

package {{.DistFilePkgName}}

{{if ne .FromPkgPathImport ""}}import "{{.FromPkgPathImport}}"{{end}}

func {{.FuncNameToDTO}}(src {{.FromStructName}}) {{.ToStructName}} {
	return {{.ToStructName}}{
		{{range .Fields}}{{.ToField.Name}}: src.{{.FromField}},
		{{end}}
	}
}

func {{.FuncNameToStruct}}(src {{.ToStructName}}) {{.FromStructName}} {
	return {{.FromStructName}}{
		{{range .Fields}}{{.FromField}}: src.{{.ToField.Name}},
		{{end}}
	}
}
`

func (data TemplateData) GenerateCode(output io.Writer) error {
	t := template.Must(template.New("morph").Parse(tmpl))
	err := t.Execute(output, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func (t FromStructType) FileName() string {
	return fmt.Sprintf("morph_%s.go", strings.ToLower(t.Name))
}
