package differentfiles

//go:generate go run ../../cmd/structmorph/structmorph.go --src=differentfiles.Person --dst=differentfiles.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
