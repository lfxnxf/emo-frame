package logging

var DefaultKit Kit

type Kit interface {
	// Access log
	A() *Logger
	// Error log
	E() *Logger
	// info log
	I() *Logger
	// debug log
	D() *Logger
	// sql log
	S() *Logger
	// business log
	B() *Logger
}

type kit struct {
	a, e, i, d, s, b *Logger
}

func NewKit(a, e, i, d, s, b *Logger) Kit {
	return kit{
		a: a,
		e: e,
		i: i,
		d: d,
		s: s,
		b: b,
	}
}

func (c kit) A() *Logger {
	return c.a
}

func (c kit) E() *Logger {
	return c.e
}

func (c kit) I() *Logger {
	return c.i
}

func (c kit) D() *Logger {
	return c.d
}

func (c kit) S() *Logger {
	return c.s
}

func (c kit) B() *Logger {
	return c.b
}
