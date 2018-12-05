package hosting

import (
	"context"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
)

type Server struct {
	context context.Context
	cancel  context.CancelFunc
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Serve() {
	cts, s.cancel = context.WithCancel(context.Background())

	host, err := libp2p.New(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("hello world. host id is %s\n", host.ID())
}

func (s *Server) Quit() {
	s.cancel()
}
