package main

import (
	"fmt"
	"internal/mock"
	"runtime"
	"syscall"
)

func main() {
	filename, ok := mock.RuntimeCallerSupport0()
	fmt.Println(ok)
	fmt.Println(filename)
	filename, ok = mock.RuntimeCallerSupport1()
	fmt.Println(ok)
	fmt.Println(filename)
	fmt.Println(runtime.GOROOT())
	fmt.Println(syscall.Environ())
	fmt.Println(syscall.Getenv("DNDNR_ROOT"))
	fmt.Println(syscall.Getenv("foo"))
	//var env := syscall.Environ()
	//fmt.Println(env["DNDNR_ROOT"])

}
