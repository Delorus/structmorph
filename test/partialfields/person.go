package partialfields

type Person struct {
	Name string
	Sex  string
	Age  int
}

//go:generate go run ../../cmd/structmorph/structmorph.go --src=partialfields.Person --dst=partialfields.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
