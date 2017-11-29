package network

import "fmt"
import "log"
import "net"

type Peer struct {
	port int
	conn net.Conn
	live bool
}

func NewPeer(port int) *Peer {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Print("outgoing connection error: ", err)
		return &Peer{port: port, live: false}
	}
		
	fmt.Fprintf(conn, "hello\n")
	return &Peer{port: port, conn: conn, live: true}
}
