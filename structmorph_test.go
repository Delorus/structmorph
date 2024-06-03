package structmorph

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractStructFields(t *testing.T) {
	// Setup
	src := `package main

    type Person struct {
        Name string
        Age  int
        Sex  bool
    }

    type PersonDTO struct {
        Name string
        Age  int
        Sex  bool
    }`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	assert.NoError(t, err, "Error parsing file")

	expectedFromFields := map[string]string{
		"Name": "string",
		"Age":  "int",
		"Sex":  "bool",
	}

	expectedToFields := map[string]string{
		"Name": "string",
		"Age":  "int",
		"Sex":  "bool",
	}

	// When
	toFields := extractStructFields(f, "Person")
	fromFields := extractStructFields(f, "PersonDTO")

	// Then
	assert.Equal(t, expectedFromFields, fromFields, "fromFields do not match")
	assert.Equal(t, expectedToFields, toFields, "toFields do not match")
}

func Test_generateCode(t *testing.T) {
	// Setup
	data := TemplateData{
		FuncNameToDTO:    "ConvertToPerson",
		FuncNameToStruct: "ConvertToPersonDTO",
		FromPkg:          "main",
		FromPkgPath:      "",
		From:             "Person",
		To:               "PersonDTO",
		Fields: []FieldMapping{
			FieldMapping{
				FromField: "Name",
				ToField:   "Name",
			},
			FieldMapping{
				FromField: "Age",
				ToField:   "Age",
			},
			FieldMapping{
				FromField: "Sex",
				ToField:   "Sex",
			},
		},
	}

	// When
	err := generateCode(data, "Person")

	// Then
	assert.NoError(t, err, "Error generating code")
}

func Test_mapFields(t *testing.T) {
	// Setup
	toFields := map[string]string{
		"Name": "string",
		"Age":  "int",
		"Sex":  "bool",
	}

	expected := []FieldMapping{
		FieldMapping{
			FromField: "Name",
			ToField:   "Name",
		},
		FieldMapping{
			FromField: "Age",
			ToField:   "Age",
		},
		FieldMapping{
			FromField: "Sex",
			ToField:   "Sex",
		},
	}

	// When
	actual := mapFields(toFields)

	// Then
	assert.Contains(t, actual, expected[0], "mapFields does not contain expected field")
	assert.Contains(t, actual, expected[1], "mapFields does not contain expected field")
	assert.Contains(t, actual, expected[2], "mapFields does not contain expected field")
}
