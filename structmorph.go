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
	Fields     map[string]string
}

type DstFieldType struct {
	Name     string
	Type     string
	SrcField string
}

type FieldMapping struct {
	SrcField string
	DstField DstFieldType
}

func CreateMapping(srcStruct SrcStructType, dstStruct DstStructType) ([]FieldMapping, error) {
	var fields []FieldMapping
	for _, dstField := range dstStruct.Fields {
		srcField := dstField.SrcField
		srcFieldType, ok := srcStruct.Fields[srcField]
		if !ok || srcFieldType != dstField.Type {
			return nil, fmt.Errorf("field not found or type mismatch, field: %s, type: %+v, struct: %s", srcField, dstField, srcStruct.Name)
		}
		fields = append(fields, FieldMapping{
			SrcField: srcField,
			DstField: dstField,
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
	FuncNameToDTO    string
	FuncNameToStruct string
	SrcPkgPathImport string
	SrcStructName    string
	DistFilePkgName  string
	DstStructName    string
	Fields           []FieldMapping
}

const tmpl = `// Code generated by structmorph; DO NOT EDIT.

package {{.DistFilePkgName}}

{{if ne .SrcPkgPathImport ""}}import "{{.SrcPkgPathImport}}"{{end}}

func {{.FuncNameToDTO}}(src {{.SrcStructName}}) {{.DstStructName}} {
	return {{.DstStructName}}{
		{{range .Fields}}{{.DstField.Name}}: src.{{.SrcField}},
		{{end}}
	}
}

func {{.FuncNameToStruct}}(src {{.DstStructName}}) {{.SrcStructName}} {
	return {{.SrcStructName}}{
		{{range .Fields}}{{.SrcField}}: src.{{.DstField.Name}},
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

func (t SrcStructType) FileName() string {
	return fmt.Sprintf("morph_%s.go", strings.ToLower(t.Name))
}
