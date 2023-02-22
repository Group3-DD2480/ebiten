uniform mat4 U0[4];

mat4[2] F0(void);
mat4[2] F1(void);

mat4[2] F0(void) {
	mat4 l0[2];
	l0[0] = mat4(0);
	l0[1] = mat4(0);
	return l0;
}

mat4[2] F1(void) {
	mat4 l0[2];
	l0[0] = mat4(0);
	l0[1] = mat4(0);
	mat4 l1[2];
	l1[0] = mat4(0);
	l1[1] = mat4(0);
	(l0)[0] = mat4(1.0);
	l1[0] = l0[0];
	l1[1] = l0[1];
	(l1)[1] = mat4(2.0);
	return l1;
}
