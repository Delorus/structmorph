package main

//go:generate structmorph --from=main.Person --to=main.PersonDTO
type Person struct {
	Name string
	Age  int
	Sex  bool
}

type PersonDTO struct {
	Name string
	Sex  bool
}
