uniform mat2 U0[4];

mat2[2] F0(void);
mat2[2] F1(void);

mat2[2] F0(void) {
	mat2 l0[2];
	l0[0] = mat2(0);
	l0[1] = mat2(0);
	return l0;
}

mat2[2] F1(void) {
	mat2 l0[2];
	l0[0] = mat2(0);
	l0[1] = mat2(0);
	mat2 l1[2];
	l1[0] = mat2(0);
	l1[1] = mat2(0);
	(l0)[0] = mat2(1.0);
	l1[0] = l0[0];
	l1[1] = l0[1];
	(l1)[1] = mat2(2.0);
	return l1;
}
