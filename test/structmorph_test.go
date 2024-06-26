package structmorph__test

import (
	"structmorph"
	"testing"
)

func TestGenerateManual(t *testing.T) {
	t.Skip("Skip manual test")
	// you can generate in manual mode for debug purposes
	err := structmorph.Generate("second.Person", "first.PersonDTO", structmorph.WithProjectRoot("testdata/differentpkg"))
	if err != nil {
		t.Fatal(err)
	}
}
