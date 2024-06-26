package samepackage

type Person struct {
	Name string
	Age  int
}

//go:generate go run ../../cmd/structmorph/structmorph.go --src=samepackage.Person --dst=samepackage.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
