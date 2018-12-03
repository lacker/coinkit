package hosting

import ()

type Server struct {
	port int
}

func NewServer(port int) *Server {
	return &Server{port: port}
}
