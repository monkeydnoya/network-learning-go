package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func server(fd int, c chan int) {
	for {
		/* TODO: Show connected client socket addr */
		nfd, cliaddr, err := syscall.Accept(fd)
		if err != nil {
			println("failed to accept connection: ", err.Error())
			c <- -1
			return
		}

		switch cla := cliaddr.(type) {
		case *syscall.SockaddrInet4:
			fmt.Printf("client: %s:%d, fd: %d\n", cla.Addr, cla.Port, nfd)
		default:
			fmt.Printf("client socket type not supported: %v, fd: %d\n", cla, nfd)
		}

		go func(nfd int) {
			buff := make([]byte, 1024)

			for {
				n, err := syscall.Read(nfd, buff)
				if err != nil {
					println("failed to read data: ", err.Error())
					syscall.Close(nfd)
					return
				}

				/*
					When client terminates connection 0 bytes will be returned whitout any error
				*/
				println("read bytes: ", n)

				if n > 0 {
					_, err = syscall.Write(nfd, buff[:n])
					if err != nil {
						println("failed to write data: ", err.Error())
						syscall.Close(nfd)
						return
					}
				}

				syscall.Close(nfd)
				return
			}
		}(nfd)
	}
}

func main() {
	PrintHostByteOrdering()

	/* proto - Protocol type (TCP, UDP, etc). 0 means chose protocol type automatically */
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	defer syscall.Close(fd)

	/* TODO: Read abot time_wait after socket closed */
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		panic(err)
	}

	/* TODO: Network to host*/
	sockaddr := syscall.SockaddrInet4{
		Port: 8000,
		Addr: [4]byte{0, 0, 0, 0},
	}

	if err := syscall.Bind(fd, &sockaddr); err != nil {
		panic(err)
	}

	if err := syscall.Listen(fd, 0); err != nil {
		panic(err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	doneCh := make(chan int, 1)
	go server(fd, doneCh)

outer:
	for {
		select {
		case signal := <-sigchan:
			println("accepted signal: ", signal)
			break outer
		case code := <-doneCh:
			println("done: ", code)
			break outer
		}
	}

	close(sigchan)
	close(doneCh)

	println("exiting")
}
