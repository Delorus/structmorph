package partialfields

type Person struct {
	Name string
	Sex  string
	Age  int
}

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=partialfields.Person --to=partialfields.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
