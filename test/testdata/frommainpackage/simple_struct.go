package main

type Person struct {
	Name string
	Age  int
	Sex  bool
}

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=main.Person --to=main.PersonDTO
type PersonDTO struct {
	Name string
	Sex  bool
}
