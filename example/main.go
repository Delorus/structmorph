package main

import "structmorph"

func main() {
	err := structmorph.Generate("main.Person", "main.PersonDTO")
	if err != nil {
		panic(err)
	}
}
