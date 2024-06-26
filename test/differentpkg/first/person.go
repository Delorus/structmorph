package first

//go:generate go run ../../../cmd/structmorph/structmorph.go --src=second.Person --dst=first.PersonDTO --root=../.
type PersonDTO struct {
	Name string
	Age  int
}
