varying vec2 V0;
varying vec4 V1;

vec4 F0(in vec4 l0, in vec2 l1, in vec4 l2);

vec4 F0(in vec4 l0, in vec2 l1, in vec4 l2) {
	if (false) {
		l0 = (l0) / (5.0);
	}
	if (true) {
		l0 = (l0) / (15.0);
	}
	return l0;
}