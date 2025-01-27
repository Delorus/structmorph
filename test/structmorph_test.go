package structmorph__test

import (
	"structmorph"
	"structmorph/test/allsupportedtypes"
	"structmorph/test/customfieldname"
	"structmorph/test/partialfields"
	"structmorph/test/pointers"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateManual(t *testing.T) {
	t.Skip("Skip manual test")
	// you can generate in manual mode for debug purposes
	err := structmorph.Generate("pointers.Organization", "pointers.OrganizationDTO", structmorph.WithProjectRoot("pointers"))
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

func TestGenerate__allsupportedtypes(t *testing.T) {
	// Setup
	tp := allsupportedtypes.Type{}
	err := faker.FakeData(&tp, options.WithFieldsToIgnore("InterfaceField"))
	require.NoError(t, err)
	tp.InterfaceField = faker.Word()

	// When
	tpDTO := allsupportedtypes.ConvertToTypeDTO(tp)
	convertedType := allsupportedtypes.ConvertToType(tpDTO)

	// Then
	assert.Equal(t, tp, convertedType)
}

func TestGenerate__pointers(t *testing.T) {
	// Setup
	org := pointers.Organization{}
	err := faker.FakeData(&org)
	require.NoError(t, err)

	// When
	orgDTO := pointers.ConvertToOrganizationDTO(org)
	convertedOrg := pointers.ConvertToOrganization(orgDTO)

	// Then
	assert.Equal(t, org.Title, *orgDTO.Title)
	assert.Equal(t, *org.Description, orgDTO.Description)
	assert.Equal(t, org.Priority, orgDTO.Priority)

	assert.Equal(t, org.Title, convertedOrg.Title)
	assert.Equal(t, *org.Description, *convertedOrg.Description)
	assert.Equal(t, org.Priority, convertedOrg.Priority)
}

func TestGenerate__pointers__nilInSource(t *testing.T) {
	// Setup
	org := pointers.Organization{
		Title:       "Title",
		Description: nil,
	}

	// When
	orgDTO := pointers.ConvertToOrganizationDTO(org)
	convertedOrg := pointers.ConvertToOrganization(orgDTO)

	// Then
	assert.Equal(t, org.Title, *orgDTO.Title)
	assert.Equal(t, "", orgDTO.Description)
	assert.Equal(t, org.Priority, orgDTO.Priority)

	assert.Equal(t, org.Title, convertedOrg.Title)
	assert.Nil(t, convertedOrg.Description)
	assert.Equal(t, org.Priority, convertedOrg.Priority)
}

func TestGenerate__pointers__emptyToPointer(t *testing.T) {
	// Setup
	description := "description"
	org := pointers.Organization{
		Title:       "",
		Description: &description,
	}

	// When
	orgDTO := pointers.ConvertToOrganizationDTO(org)
	convertedOrg := pointers.ConvertToOrganization(orgDTO)

	// Then
	assert.Nil(t, orgDTO.Title)
	assert.Equal(t, *org.Description, orgDTO.Description)
	assert.Equal(t, org.Priority, orgDTO.Priority)

	assert.Equal(t, org.Title, convertedOrg.Title)
	assert.Equal(t, *org.Description, *convertedOrg.Description)
	assert.Equal(t, org.Priority, convertedOrg.Priority)
}
