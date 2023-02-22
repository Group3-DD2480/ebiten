package main

var Array [4]mat4

func Foo() [2]mat4 {
	var x [2]mat4
	return x
}

func Bar() [2]mat4 {
	x := [2]mat4{mat4(1)}
	x[1] = mat4(2)
	return x
}
