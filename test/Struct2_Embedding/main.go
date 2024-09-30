package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

type WoodBuilder struct {
	person Person
}

func main() {
	builder := WoodBuilder{person: Person{Name: "Denis", Age: 18}}
	fmt.Println(builder.person)
}
