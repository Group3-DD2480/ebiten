package main

const foo = 10
const bar = -10

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	position /= foo*foo - foo/bar
	return position
}
