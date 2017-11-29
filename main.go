package main

import "fmt"
import "log"
import "net"
import "os"
import "strconv"

// Handles an incoming connection
func handleConnection(conn net.Conn) {
	log.Printf("handling a connection, by closing it")
	conn.Close()
}

func listen(port int) {
	log.Printf("listening on port %d", port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("connection error: ", err)
		}
		go handleConnection(conn)
	}
}

func main() {
	// Usage: go run main.go <i> where i is in [0, 1, 2, 3]
	if len(os.Args) < 2 {
		log.Fatal("Use an argument with a numerical id.")
	}
	id, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if id < 0 || id > 3 {
		log.Fatalf("invalid id: %d", id)
	}
	
	log.Printf("server %d starting up", id)

	port := 9000 + id
	go listen(port)

	// TODO: dont just end the program
	log.Printf("end of main")
}
