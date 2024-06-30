package pointers

type Organization struct {
	Title       string
	Description *string
	Priority    float64
}

//go:generate go run ../../cmd/structmorph/structmorph.go --src=pointers.Organization --dst=pointers.OrganizationDTO
type OrganizationDTO struct {
	Title       *string
	Description string
	Priority    float64
}
