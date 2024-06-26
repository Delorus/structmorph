package main

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=another.Person --to=main.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
