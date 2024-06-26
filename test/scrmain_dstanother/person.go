package main

//go:generate go run ../../cmd/structmorph/structmorph.go --src=another.Person --dst=main.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
