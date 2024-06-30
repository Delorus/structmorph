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

func Generate(src, dst string, opts ...GenerationConfigOption) error {
	cfg := DefaultGenerationConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	parser := cfg.NewParser()

	srcStructName, err := ParseStructName(src)
	if err != nil {
		return err
	}
	dstStructName, err := ParseStructName(dst)
	if err != nil {
		return err
	}

	srcStruct, err := parser.FindAndParseStructSrc(srcStructName)
	if err != nil {
		return err
	}
	slog.Info("Found and parsed struct", slog.Any("struct", srcStruct))

	dstStruct, err := parser.FindAndParseStructDst(dstStructName)
	if err != nil {
		return err
	}
	slog.Info("Found and parsed struct", slog.Any("struct", dstStruct))

	data, err := CreateTemplateData(srcStruct, dstStruct)
	if err != nil {
		return fmt.Errorf("error creating template data: %w", err)
	}

	buff := &bytes.Buffer{}
	err = data.GenerateCode(buff)
	if err != nil {
		return fmt.Errorf("error generating code: %w", err)
	}

	fileName := filepath.Join(dstStruct.Filepath(), srcStruct.FileName())
	err = FormatAndWrite(buff, fileName)
	if err != nil {
		return fmt.Errorf("error formatting and writing code: %w", err)
	}

	slog.Info("Generated and formatted code", "file", fileName)
	return nil
}

func CreateTemplateData(srcStruct SrcStructType, dstStruct DstStructType) (TemplateData, error) {
	data := TemplateData{
		FuncNameToDTO:    fmt.Sprintf("ConvertTo%s", dstStruct.Name),
		FuncNameToStruct: fmt.Sprintf("ConvertTo%s", srcStruct.Name),
		SrcStructName:    srcStruct.Name,
		DstStructName:    dstStruct.Name,
		DistFilePkgName:  dstStruct.Package,
	}

	data.SrcStructName = resolveSrcStructName(srcStruct, dstStruct)
	data.SrcPkgPathImport = resolveSrcStructImport(srcStruct, dstStruct)

	fields, err := CreateMapping(srcStruct, dstStruct)
	if err != nil {
		return data, err
	}
	data.Fields = fields

	err = CreateMods(&data)
	if err != nil {
		return data, fmt.Errorf("error creating mods: %w", err)
	}

	return data, nil
}

func resolveSrcStructName(srcStruct SrcStructType, dstStruct DstStructType) string {
	if srcStruct.Package == dstStruct.Package {
		return srcStruct.Name
	}
	return fmt.Sprintf("%s.%s", srcStruct.Package, srcStruct.Name)
}

func resolveSrcStructImport(srcStruct SrcStructType, dstStruct DstStructType) string {
	if srcStruct.Package == dstStruct.Package {
		return ""
	}
	return srcStruct.ImportPath
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
		return StructName{}, fmt.Errorf("invalid format for struct name. Expected 'package.StructName'")
	}
	return StructName{
		Package: parts[0],
		Name:    parts[1],
	}, nil
}

type DstStructType struct {
	StructName
	Fields   []DstFieldType
	FilePath string
}

func (s *DstStructType) Filepath() string {
	return s.FilePath
}

type SrcStructType struct {
	StructName
	ImportPath string
	Fields     map[string]SrcFieldType
}

type DstFieldType struct {
	FieldType
	SrcField string
}

type SrcFieldType struct {
	FieldType
}

type FieldType struct {
	Name           string
	Type           FieldTypeType
	OverriddenName string
}

type FieldTypeType struct {
	Name      string
	IsPointer bool
}

type FieldMapping struct {
	SrcField SrcFieldType
	DstField DstFieldType
}

func CreateMapping(srcStruct SrcStructType, dstStruct DstStructType) ([]FieldMapping, error) {
	var fields []FieldMapping
	for _, dstField := range dstStruct.Fields {
		srcField := dstField.SrcField
		srcFieldType, ok := srcStruct.Fields[srcField]
		if !ok {
			return nil, fmt.Errorf("field not found, field: %s, struct: %s", srcField, srcStruct.Name)
		}
		if srcFieldType.Type.Name != dstField.Type.Name {
			return nil, fmt.Errorf("field type mismatch, field: %s, src: %s, dst: %s", srcField, srcFieldType.Type.Name, dstField.Type.Name)
		}
		fields = append(fields, FieldMapping{
			SrcField: srcFieldType,
			DstField: dstField,
		})
	}

	return fields, nil
}

var tmplDeref = template.Must(template.New("deref").Parse(`
var {{.OverriddenName}} {{.Type.Name}}
if src.{{.Name}} != nil {
	{{.OverriddenName}} = *src.{{.Name}}
}
`))

var tmplRef = template.Must(template.New("ref").Parse(`
var {{.OverriddenName}} *{{.Type.Name}}
if src.{{.Name}} != *new({{.Type.Name}}) {
	{{.OverriddenName}} = &src.{{.Name}}
}
`))

func CreateMods(t *TemplateData) error {
	for i, field := range t.Fields {
		if isFromPtrToValue(field.SrcField, field.DstField) {
			t.Fields[i].SrcField.OverriddenName = fmt.Sprintf("__synthetic__%s", field.SrcField.Name)
			deref, err := renderDeref(t.Fields[i].SrcField.FieldType)
			if err != nil {
				return fmt.Errorf("error rendering deref: %w", err)
			}
			t.ModsToDTO = append(t.ModsToDTO, deref)

			t.Fields[i].DstField.OverriddenName = fmt.Sprintf("__synthetic__%s", field.DstField.Name)
			ref, err := renderRef(t.Fields[i].DstField.FieldType)
			if err != nil {
				return fmt.Errorf("error rendering ref: %w", err)
			}
			t.ModsToStruct = append(t.ModsToStruct, ref)

		}
		if isFromValueToPtr(field.SrcField, field.DstField) {
			t.Fields[i].SrcField.OverriddenName = fmt.Sprintf("__synthetic__%s", field.SrcField.Name)
			ref, err := renderRef(t.Fields[i].SrcField.FieldType)
			if err != nil {
				return fmt.Errorf("error rendering ref: %w", err)
			}
			t.ModsToDTO = append(t.ModsToDTO, ref)

			t.Fields[i].DstField.OverriddenName = fmt.Sprintf("__synthetic__%s", field.DstField.Name)
			deref, err := renderDeref(t.Fields[i].DstField.FieldType)
			if err != nil {
				return fmt.Errorf("error rendering deref: %w", err)
			}
			t.ModsToStruct = append(t.ModsToStruct, deref)
		}
	}

	return nil
}

func renderDeref(field FieldType) (string, error) {
	buff := &bytes.Buffer{}
	err := tmplDeref.Execute(buff, field)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buff.String(), nil
}

func renderRef(field FieldType) (string, error) {
	buff := &bytes.Buffer{}
	err := tmplRef.Execute(buff, field)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buff.String(), nil
}

func isFromPtrToValue(srcField SrcFieldType, dstField DstFieldType) bool {
	return srcField.Type.IsPointer && !dstField.Type.IsPointer
}

func isFromValueToPtr(srcField SrcFieldType, dstField DstFieldType) bool {
	return !srcField.Type.IsPointer && dstField.Type.IsPointer
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
	FuncNameToDTO    string
	FuncNameToStruct string
	SrcPkgPathImport string
	SrcStructName    string
	DistFilePkgName  string
	DstStructName    string
	ModsToDTO        []string
	ModsToStruct     []string
	Fields           []FieldMapping
}

var tmpl = template.Must(template.New("morph").Parse(`// Code generated by structmorph; DO NOT EDIT.

package {{.DistFilePkgName}}

{{if .SrcPkgPathImport}}import "{{.SrcPkgPathImport}}"{{end}}

func {{.FuncNameToDTO}}(src {{.SrcStructName}}) {{.DstStructName}} {
	{{range .ModsToDTO -}}{{.}}{{end}}
	return {{.DstStructName}}{
		{{range .Fields}}{{.DstField.Name}}: {{with .SrcField.OverriddenName}}{{.}}{{else}}src.{{.SrcField.Name}}{{end}},
		{{end}}
	}
}

func {{.FuncNameToStruct}}(src {{.DstStructName}}) {{.SrcStructName}} {
	{{range .ModsToStruct -}}{{.}}{{end}}
	return {{.SrcStructName}}{
		{{range .Fields}}{{.SrcField.Name}}: {{with .DstField.OverriddenName}}{{.}}{{else}}src.{{.DstField.Name}}{{end}},
		{{end}}
	}
}
`))

func (data TemplateData) GenerateCode(output io.Writer) error {
	err := tmpl.Execute(output, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}

func (t SrcStructType) FileName() string {
	return fmt.Sprintf("morph_%s.go", strings.ToLower(t.Name))
}
