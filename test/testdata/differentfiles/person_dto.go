package differentfiles

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=differentfiles.Person --to=differentfiles.PersonDTO
type PersonDTO struct {
	Name string
	Age  int
}
