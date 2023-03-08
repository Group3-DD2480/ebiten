varying vec2 V0;
varying vec4 V1;

vec4 F0(in vec4 l0, in vec2 l1, in vec4 l2);

vec4 F0(in vec4 l0, in vec2 l1, in vec4 l2) {
	l0 = (l0) / (11.0);
	return l0;
}