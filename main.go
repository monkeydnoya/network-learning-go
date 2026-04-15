package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const CRLF = "\r\n"

func Router(w *Response, r *Request) {
	switch r.URI {
	case "/plaintext":
		/* Example: HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Lenght: 13\r\n\r\nHello, World! */
		message := "plaintext"

		w.Arena.CopyString(r.Proto)
		w.Arena.CopyString(" ")

		w.Arena.CopyString("200 OK")
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString("Content-Type: text/plain")
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString("Content-Length: ")
		w.Arena.CopyString(strconv.Itoa(len(message)))
		w.Arena.CopyString(CRLF)
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString(message)
	case "/html":
		html := `<!DOCTYPE html>
<html>
<head>
<title>My Simple Page</title>
</head>
<body>
<h1>Hello, World!</h1>
<p>This is a simple HTML page.</p>
</body>
</html>`

		w.Arena.CopyString(r.Proto)
		w.Arena.CopyString(" ")

		w.Arena.CopyString("200 OK")
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString("Content-Type: text/html")
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString("Content-Length: ")
		w.Arena.CopyString(strconv.Itoa(len(html)))
		w.Arena.CopyString(CRLF)
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString(html)
	default:
		html := `<!DOCTYPE html>
<html>
<head><title>Page not found</title></head>
<body><h1>404</h1><p>Page not found</p></body>
</html>
`

		w.Arena.CopyString(r.Proto)
		w.Arena.CopyString(" ")

		w.Arena.CopyString("404 Not Found")
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString("Content-Type: text/html")
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString("Content-Length: ")
		w.Arena.CopyString(strconv.Itoa(len(html)))
		w.Arena.CopyString(CRLF)
		w.Arena.CopyString(CRLF)

		w.Arena.CopyString(html)
	}
}

func Serve(fd int, c chan int) {
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
			fmt.Printf("client:%s:%d, fd:%d\n", cla.Addr, cla.Port, nfd)
		default:
			fmt.Printf("client socket addr type not supported: %v, fd: %d\n", cla, nfd)
		}

		go func(nfd int) {
			defer syscall.Close(nfd)

			var pos int

			/* TODO: Buffer overflow */
			buff := make([]byte, 1024)

			/* TODO: Handle errors */
			for {
				n, err := syscall.Read(nfd, buff[pos:])
				if err != nil {
					println("failed to read data: ", err.Error())
					break
				}

				if n == 0 {
					println("EOF")
					break
				}

				pos += n
				rEnd := strings.Index(*(*string)(unsafe.Pointer(&buff)), "\r\n\r\n")
				if rEnd != -1 {
					break
				}
			}

			var r Request
			rString := *(*string)(unsafe.Pointer(&buff))

			startLineEnd := strings.IndexByte(rString, '\r')
			methodEnd := strings.IndexByte(rString[:startLineEnd], ' ')

			r.Method = r.Arena.CopyString(rString[:methodEnd])

			uriEnd := strings.IndexByte(rString[methodEnd+1:startLineEnd], ' ') + len(r.Method) + 1
			r.URI = r.Arena.CopyString(rString[methodEnd+1 : uriEnd])

			r.Proto = r.Arena.CopyString(rString[uriEnd+1 : startLineEnd])

			var w Response
			Router(&w, &r)

			fmt.Println(string(w.Arena.Buf))

			/* Write response to socket */
			_, err := syscall.Write(nfd, w.Arena.Buf)
			if err != nil {
				println("failed to write response: ", err.Error())
				return
			}

			clear(buff)
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

	/* TODO: Read abot TIME_WAIT after socket closed */
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
	go Serve(fd, doneCh)

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
