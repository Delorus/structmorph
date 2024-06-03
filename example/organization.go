package main

//go:generate structmorph --from=domain.Organization --to=main.OrganizationDTO
type OrganizationDTO struct {
	Title       string
	Description string
	Priority    float64
}
