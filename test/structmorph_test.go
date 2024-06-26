package structmorph__test

import (
	"structmorph"
	"structmorph/test/testdata/customfieldname"
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

func TestGenerate__customfieldname(t *testing.T) {
	// Setup
	org := customfieldname.Organization{}
	err := faker.FakeData(&org)
	require.NoError(t, err)

	// When
	orgDTO := customfieldname.ConvertToOrganizationDTO(org)
	convertedOrg := customfieldname.ConvertToOrganization(orgDTO)

	// Then
	assert.Equal(t, org.Title, orgDTO.Title)
	assert.Equal(t, org.Description, orgDTO.Description)
	assert.Equal(t, org.Priority, orgDTO.Priority)
	assert.Equal(t, org.EmployeesCount, orgDTO.TeamSize)

	assert.Equal(t, org.Title, convertedOrg.Title)
	assert.Equal(t, org.Description, convertedOrg.Description)
	assert.Equal(t, org.Priority, convertedOrg.Priority)
	assert.Equal(t, org.EmployeesCount, convertedOrg.EmployeesCount)
}
