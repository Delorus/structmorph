package structmorph__test

import (
	"structmorph"
	"testing"
)

func TestSimpleStruct(t *testing.T) {
	err := structmorph.Generate("main.Person", "main.PersonDTO")
	if err != nil {
		t.Fatal(err)
	}
}
