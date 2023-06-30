package main

import "net"

func main() {
	lis, err := net.Listen("tcp", "10.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	for {
		conn, _ := lis.Accept()
		conn.Write([]byte("Hello, world!"))
		conn.Close()
	}
}
