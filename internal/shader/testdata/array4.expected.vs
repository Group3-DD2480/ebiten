uniform mat3 U0[4];

mat3[2] F0(void);
mat3[2] F1(void);

mat3[2] F0(void) {
	mat3 l0[2];
	l0[0] = mat3(0);
	l0[1] = mat3(0);
	return l0;
}

mat3[2] F1(void) {
	mat3 l0[2];
	l0[0] = mat3(0);
	l0[1] = mat3(0);
	mat3 l1[2];
	l1[0] = mat3(0);
	l1[1] = mat3(0);
	(l0)[0] = mat3(1.0);
	l1[0] = l0[0];
	l1[1] = l0[1];
	(l1)[1] = mat3(2.0);
	return l1;
}
