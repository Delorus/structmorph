package main

//go:generate go run ../cmd/structmorph/structmorph.go --from=main.Person --to=main.PersonDTO
type Person struct {
	Name string
	Age  int
	Sex  bool
}

type PersonDTO struct {
	Name string
	Sex  bool
}
