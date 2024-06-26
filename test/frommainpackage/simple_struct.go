package main

type Person struct {
	Name string
	Age  int
	Sex  bool
}

//go:generate go run ../../cmd/structmorph/structmorph.go --src=main.Person --dst=main.PersonDTO
type PersonDTO struct {
	Name string
	Sex  bool
}
