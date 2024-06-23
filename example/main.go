package main

import "structmorph"

func main() {
	err := structmorph.Generate("domain.Organization", "main.OrganizationDTO")
	if err != nil {
		panic(err)
	}
}
