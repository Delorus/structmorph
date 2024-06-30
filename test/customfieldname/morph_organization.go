// Code generated by structmorph; DO NOT EDIT.

package customfieldname

func ConvertToOrganizationDTO(src Organization) OrganizationDTO {

	return OrganizationDTO{
		Title:       src.Title,
		Description: src.Description,
		Priority:    src.Priority,
		TeamSize:    src.EmployeesCount,
	}
}

func ConvertToOrganization(src OrganizationDTO) Organization {

	return Organization{
		Title:          src.Title,
		Description:    src.Description,
		Priority:       src.Priority,
		EmployeesCount: src.TeamSize,
	}
}
