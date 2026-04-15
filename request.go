package main

type Request struct {
	Method string
	URI    string
	Proto  string

	Arena Arena
}
