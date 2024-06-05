package GeeCache

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return CloneByte(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func CloneByte(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
