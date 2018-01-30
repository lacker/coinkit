package network

import (
	"log"
	"testing"
)

func makeServers() []*Server {
	answer := []*Server{}
	for i := 0; i <= 3; i++ {
		config := NewLocalConfig(i)
		server := NewServer(config)
		server.ServeInBackground()
		answer = append(answer, server)
	}
	return answer
}

func stopServers(servers []*Server) {
	for _, server := range servers {
		server.Stop()
	}
}

func TestStartStop(t *testing.T) {
	servers := makeServers()
	stopServers(servers)
	moreServers := makeServers()
	stopServers(moreServers)
}

