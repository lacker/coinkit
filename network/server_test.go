package network

import (
	"testing"
)

func TestBasicNetwork(t *testing.T) {
	c0 := NewLocalConfig(0)
	c1 := NewLocalConfig(1)
	c2 := NewLocalConfig(2)
	s0 := NewServer(c0)
	s1 := NewServer(c1)
	s2 := NewServer(c2)
	go s0.ServeForever()
	go s1.ServeForever()
	go s2.ServeForever()
}

func TestNewServerCreatesSufficientPeers(t *testing.T) {
	c0 := NewLocalConfig(0)
	s0 := NewServer(c0)

	if (len(s0.peers) != NumPeers - 1) {
		t.Errorf("Didn't create the right number of peers %f %f", len(s0.peers), NumPeers - 1);
	}
}
