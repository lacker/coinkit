package network

import (
	"io"
	"testing"
	"time"
	"fmt"
	"coinkit/util"
)

type FakeMessage struct {
	Number int
}

func (m *FakeMessage) Slot() int {
	return 0
}

func (m *FakeMessage) MessageType() string {
	return "Fake"
}

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

func TestNewServerFailsIfPortTaken(t *testing.T) {
	s0 := NewServer(NewLocalConfig(0))
	s1 := NewServer(NewLocalConfig(0))

	go s0.ServeForever()
	err := s1.ServeForever()
	if (err == nil) {
		t.Errorf("Didn't error out when port is already in use")
	}
}

func TestServerOkayWithFakeWellFormattedMessage(t *testing.T) {
	s0 := NewServer(NewLocalConfig(0))

	m := &FakeMessage{Number: 4}
	kp := util.NewKeyPairFromSecretPhrase("foo")
	sm := util.NewSignedMessage(kp, m)

	fakeRequest := &Request {
		Message: sm,
		Response: nil,
	}

	go s0.ServeForever()
	s0.requests <- fakeRequest
	// This hits the right code path but it feels like we ought to have a real assertion here
}

func ResetConnectionAndSendString(c *Client, s string) {
	c.connected = false
	c.connect()
	c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	fmt.Fprintf(c.conn, s)
}

func TestServerOkayWithMalformedMessage(t *testing.T) {
	s := NewServer(NewLocalConfig(0))
	s.ServeForever()

	c := NewClient(s.port)

	ResetConnectionAndSendString(c, "Hello, I am sending you garbage.\n\n\n")
	_, err := util.ReadSignedMessage(c.conn)
	if (err != io.EOF) {
		t.Errorf("Didn't get disconnected after a malformed message")
	}

	ResetConnectionAndSendString(c, "a:b:c:d\n")
	_, err2 := util.ReadSignedMessage(c.conn)
	if (err2 != io.EOF) {
		t.Errorf("Didn't get disconnected after a malformed message")
	}

	goodMessage := "{ \"T\": \"N\", \"M\": { \"I\": 1 } }"
	kp := util.NewKeyPair()
	ResetConnectionAndSendString(c,
		fmt.Sprintf("e:%s:%s:%s\n", kp.PublicKey(), "notRealSignature", goodMessage))

	_, err3 := util.ReadSignedMessage(c.conn)
	if (err3 != io.EOF) {
		t.Errorf("Didn't get disconnected after a malformed message")
	}

	// Let's test that the server is still in good shape and can process a good message
	ResetConnectionAndSendString(c, fmt.Sprintf("e:%s:%s:%s\n", kp.PublicKey(), kp.Sign(goodMessage), goodMessage))
	_, err4 := util.ReadSignedMessage(c.conn)

	if (err4 != nil) {
		t.Errorf("Couldn't get a response after the good message")
	}
}
