package main

import "unsafe"

type Arena struct {
	Buf  []byte
	Tail int
}

func (a *Arena) GetSlice(n int) []byte {
	/* TODO: Grow arena logic */
	if (len(a.Buf) + n) > cap(a.Buf) {
		old := a.Buf
		if (len(a.Buf) + n) > 2*cap(a.Buf) {
			a.Buf = make([]byte, len(a.Buf)+n+2*cap(a.Buf))
		} else {
			a.Buf = make([]byte, 2*cap(a.Buf))
		}

		copy(a.Buf, old)
	}

	sl := a.Buf[a.Tail : a.Tail+n]
	a.Tail += n

	return sl
}

func (a *Arena) Copy(src []byte) []byte {
	dst := a.GetSlice(len(src))
	copy(dst, src)

	return dst
}

func (a *Arena) CopyString(s string) string {
	dst := a.GetSlice(len(s))
	copy(dst, s)

	return *(*string)(unsafe.Pointer(&dst))
}
