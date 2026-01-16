package main

import (
	"runtime"
	"unsafe"
)

/* Print host byte ordering */
func PrintHostByteOrdering() {
	var t int16
	t = 258

	pFirstPtr := (*int8)(unsafe.Pointer(&t)) // 00000001 00000010
	pSecondPtr := (*int8)(unsafe.Pointer((uintptr)(unsafe.Pointer(&t)) + 1))

	if *pFirstPtr == 2 && *pSecondPtr == 1 {
		println("little-endian", runtime.GOOS, runtime.GOARCH)
	} else if *pFirstPtr == 1 && *pSecondPtr == 2 {
		println("big-endian", runtime.GOOS, runtime.GOARCH)
	}
}
