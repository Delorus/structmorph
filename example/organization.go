package main

//go:generate go run ../cmd/structmorph/structmorph.go --from=domain.Organization --to=main.OrganizationDTO
type OrganizationDTO struct {
	Title       string
	Description string
	Priority    float64
	TeamSize    int `morph:"EmployeesCount"`
}
