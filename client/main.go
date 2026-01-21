package main

import (
	"syscall"
)

func main() {
	servaddr := syscall.SockaddrInet4{Port: 8000, Addr: [4]byte{0, 0, 0, 0}}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}

	if err := syscall.Connect(fd, &servaddr); err != nil {
		panic(err)
	}

	buff := make([]byte, 1024)

	for {
		n, err := syscall.Read(syscall.Stdin, buff)
		if err != nil {
			panic(err)
		}

		/* Handle EOF */
		if n == 0 {
			println("EOF")
			break
		}

		if n > 0 {
			_, err := syscall.Write(fd, buff[:n])
			if err != nil {
				panic(err)
			}

			clear(buff)

			_, err = syscall.Read(fd, buff)
			if err != nil {
				panic(err)
			}

			_, err = syscall.Write(syscall.Stdout, buff)
			if err != nil {
				panic(err)
			}

			clear(buff)
		}
	}

	if err := syscall.Close(fd); err != nil {
		panic(err)
	}
}
