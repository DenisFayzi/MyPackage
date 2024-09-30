package main

import "fmt"

// Метод это Функция которая принадлежит определенному типу
//или указателю  на тип (receiver)
//

type Square struct {
	Side int
}

func (s Square) Perimetr() {
	fmt.Printf("%T,%#v\n", s, s)
	fmt.Printf("Периметр фигуры %d\n", s.Side*4)
}

func (s Square) Scale(multiplier int) {
	fmt.Printf("%T,%#v\n", s, s)
	s.Side *= multiplier

}
func (s Square) WrongScale(multiplier int) {
	fmt.Printf("%T,%#v\n", s, s)
	s.Side *= multiplier
	fmt.Printf("%T,%#v\n", s, s)
}

func main() {
	square := Square{Side: 4}
	pSquare := &square // получаем указатель на структуру

	square2 := Square{Side: 8}

	square.Perimetr()
	square2.Perimetr()
	pSquare.Perimetr()
}
