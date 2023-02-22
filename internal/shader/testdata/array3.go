package main

var Array [4]mat2

func Foo() [2]mat2 {
	var x [2]mat2
	return x
}

func Bar() [2]mat2 {
	x := [2]mat2{mat2(1)}
	x[1] = mat2(2)
	return x
}
