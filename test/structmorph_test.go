package structmorph__test

import (
	"structmorph"
	"structmorph/test/testdata/partialfields"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateManual(t *testing.T) {
	t.Skip("Skip manual test")
	// you can generate in manual mode for debug purposes
	err := structmorph.Generate("second.Person", "first.PersonDTO", structmorph.WithProjectRoot("testdata/differentpkg"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerate__partialFields(t *testing.T) {
	// Setup
	person := partialfields.Person{}
	err := faker.FakeData(&person)
	require.NoError(t, err)

	// When
	personDTO := partialfields.ConvertToPersonDTO(person)
	convertedPerson := partialfields.ConvertToPerson(personDTO)

	// Then
	assert.Equal(t, person.Name, convertedPerson.Name)
	assert.Equal(t, person.Age, convertedPerson.Age)
	assert.Empty(t, convertedPerson.Sex)
}
