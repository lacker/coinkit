package network

import "fmt"
import "log"
import "net"

func Connect(port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Print("outgoing connection error: ", err)
		return
	}
	fmt.Fprintf(conn, "hello\n")
}
