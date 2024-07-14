package main

import "fmt"

type foo struct{}

func (f *foo) bar(i int) {
	fmt.Println(i)
}

func main() {
	// p := &ohm.Addition{}
	// p.Input = "123"
	// fmt.Println(p.Expr())
}
