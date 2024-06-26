package first

//go:generate go run ../../../../cmd/structmorph/structmorph.go --from=second.Person --to=first.PersonDTO --root=../.
type PersonDTO struct {
	Name string
	Age  int
}
