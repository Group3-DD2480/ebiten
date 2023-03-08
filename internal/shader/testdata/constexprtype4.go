package main

const foo int = 10
const bar float = 5
const baz bool = true
const qux bool = false

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	if baz && qux {
		position /= foo - bar
	}
	if baz || qux {
		position /= foo + bar
	}
	return position
}
