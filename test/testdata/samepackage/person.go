package samepackage

type Person struct {
	Name string
	Age  int
}

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=samepackage.Person --to=samepackage.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
