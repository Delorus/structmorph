// Code generated by structmorph; DO NOT EDIT.

package main

import "structmorph/test/testdata/scrmain_dstanother/another"

func ConvertToPersonDTO(src another.Person) PersonDTO {
	return PersonDTO{
		Name: src.Name,
		Age:  src.Age,
	}
}

func ConvertToPerson(src PersonDTO) another.Person {
	return another.Person{
		Name: src.Name,
		Age:  src.Age,
	}
}