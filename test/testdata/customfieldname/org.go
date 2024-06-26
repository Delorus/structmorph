package customfieldname

type Organization struct {
	Title          string
	Description    string
	Priority       float64
	EmployeesCount int
}

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=customfieldname.Organization --to=customfieldname.OrganizationDTO
type OrganizationDTO struct {
	Title       string
	Description string
	Priority    float64
	TeamSize    int `morph:"EmployeesCount"`
}
