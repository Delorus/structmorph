// Code generated by structmorph; DO NOT EDIT.

package partialfields

func ConvertToPersonDTO(src Person) PersonDTO {
	return PersonDTO{
		Name: src.Name,
		Age:  src.Age,
	}
}

func ConvertToPerson(src PersonDTO) Person {
	return Person{
		Name: src.Name,
		Age:  src.Age,
	}
}
