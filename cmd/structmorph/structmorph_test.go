package main

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractStructFields(t *testing.T) {
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

	fromFields, toFields := extractStructFields(f, "Person", "PersonDTO")

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

	assert.Equal(t, expectedFromFields, fromFields, "fromFields do not match")
	assert.Equal(t, expectedToFields, toFields, "toFields do not match")
}
