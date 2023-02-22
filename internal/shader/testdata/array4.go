package main

var Array [4]mat3

func Foo() [2]mat3 {
	var x [2]mat3
	return x
}

func Bar() [2]mat3 {
	x := [2]mat3{mat3(1)}
	x[1] = mat3(2)
	return x
}
