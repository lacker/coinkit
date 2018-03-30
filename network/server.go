package network

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lacker/coinkit/currency"
	"github.com/lacker/coinkit/data"
	"github.com/lacker/coinkit/util"
)

type Server struct {
	port    int
	keyPair *util.KeyPair
	peers   []*RedialConnection
	node    *Node

	// Whenever there is a new batch of outgoing messages, it is sent to the
	// outgoing channel
	outgoing chan []*util.SignedMessage

	// Messages we are going to handle that do not require a response
	inbox chan *util.SignedMessage

	// Requests we are going to handle that do require a response
	requests chan *Request

	listener net.Listener

	// We close the currentBlock channel whenever the current block is complete
	currentBlock chan bool

	// We set shutdown to true and close the quit channel
	// when the server is shutting down
	shutdown bool
	quit     chan bool

	// A counter of how many messages we have broadcasted
	broadcasted int

	db *data.Database

	start time.Time

	// How often we send out a rebroadcast, resending our redundant data
	RebroadcastInterval time.Duration
}

func NewServer(keyPair *util.KeyPair, config *Config, db *data.Database) *Server {
	peers := []*RedialConnection{}
	inbox := make(chan *util.SignedMessage)
	for _, address := range config.PeerAddresses(keyPair) {
		peers = append(peers, NewRedialConnection(address, inbox))
	}
	qs := config.QuorumSlice()

	// At the start, all money is in the "mint" account
	mint := util.NewKeyPairFromSecretPhrase("mint")
	node := NewNodeWithMint(keyPair.PublicKey(), qs, db,
		mint.PublicKey(), currency.TotalMoney)

	return &Server{
		port:                config.Port(keyPair.PublicKey().String()),
		keyPair:             keyPair,
		peers:               peers,
		node:                node,
		outgoing:            make(chan []*util.SignedMessage, 10),
		inbox:               inbox,
		requests:            make(chan *Request),
		listener:            nil,
		shutdown:            false,
		quit:                make(chan bool),
		currentBlock:        make(chan bool),
		broadcasted:         0,
		db:                  db,
		RebroadcastInterval: time.Second,
	}
}

func (s *Server) Logf(format string, a ...interface{}) {
	util.Logf("SE", s.keyPair.PublicKey().ShortName(), format, a...)
}

// Only use for testing
func (s *Server) setBalance(user string, amount uint64) {
	s.node.queue.SetBalance(user, amount)
}

// Handles an incoming connection.
// This is likely to include many messages, all separated by endlines.
func (s *Server) handleConnection(connection net.Conn) {
	defer connection.Close()
	conn := NewBasicConnection(connection, make(chan *util.SignedMessage))

	for {
		var sm *util.SignedMessage
		select {
		case <-s.quit:
			conn.Close()
			return
		case sm = <-conn.Receive():
		}

		if sm == nil {
			return
		}

		m, ok := s.handleMessage(sm)
		if !ok {
			return
		}
		if m != nil {
			conn.Send(m)
		}
	}
}

// handleMessage will try many times for an InfoMessage, but only once for other
// messages.
// handleMessage is safe to be called from multiple threads, because it dispatches
// messages to the processing goroutine for processing.
// If we did not process the message, like if the server is shutting
// down or we are overloaded, (nil, false) is returned.
// (nil, true) means we processed the message and there is a nil response.
func (s *Server) handleMessage(sm *util.SignedMessage) (*util.SignedMessage, bool) {
	if _, ok := sm.Message().(*util.InfoMessage); ok {
		return s.retryHandleMessage(sm)
	}
	return s.handleMessageOnce(sm)
}

// handleMessageOnce is like handleMessage but explicitly only tries once.
func (s *Server) handleMessageOnce(sm *util.SignedMessage) (*util.SignedMessage, bool) {
	response := make(chan *util.SignedMessage)
	request := &Request{
		Message:  sm,
		Response: response,
	}

	// Send our request to the processing goroutine, wait for the response,
	// and return it down the connection
	s.requests <- request
	timer := time.NewTimer(time.Second)
	select {
	case m := <-response:
		return m, true
	case <-s.quit:
		return nil, false
	case <-timer.C:
		util.Logger.Fatalf("the processing goroutine got overloaded")
		return nil, false
	}
}

// retryHandleMessage is like handleMessageOnce, but it expects a non-nil response.
// If the response is nil, it waits for another block to be finalized and tries again
// when it is.
func (s *Server) retryHandleMessage(sm *util.SignedMessage) (*util.SignedMessage, bool) {
	for {
		m, ok := s.handleMessageOnce(sm)
		if !ok {
			return nil, false
		}
		if m != nil {
			return m, true
		}
		select {
		case <-s.currentBlock:
			// There's another block, so let the loop retry
		case <-s.quit:
			return nil, false
		}
	}
}

// Flushes the outgoing queue and returns the last value if there is any.
// Returns [], false if there is none
// Does not wait
func (s *Server) getOutgoing() ([]*util.SignedMessage, bool) {
	messages := []*util.SignedMessage{}
	ok := false
	for {
		select {
		case messages = <-s.outgoing:
			ok = true
		default:
			return messages, ok
		}
	}
}

// unsafeUpdateOutgoing gets the outgoing messages from our node and uses
// the outgoing channel to broadcast them.
// Since it deals with the node directly, it should only be called from the
// message-processing thread.
func (s *Server) unsafeUpdateOutgoing() {
	// Sign our messages
	out := []*util.SignedMessage{}
	for _, m := range s.node.OutgoingMessages() {
		out = append(out, util.NewSignedMessage(s.keyPair, m))
	}

	// Clear the outgoing queue
	s.getOutgoing()

	// Send our messages to the now-probably-empty queue
	s.outgoing <- out
}

// unsafeProcessMessage handles a message by interacting with the node directly.
// It should be only be called from the message-processing thread.
func (s *Server) unsafeProcessMessage(m *util.SignedMessage) *util.SignedMessage {
	prevSlot := s.node.Slot()
	message, hasResponse := s.node.Handle(m.Signer(), m.Message())
	postSlot := s.node.Slot()
	s.unsafeUpdateOutgoing()

	if postSlot != prevSlot {
		close(s.currentBlock)
		s.currentBlock = make(chan bool)
	}

	// Return the appropriate message
	if !hasResponse {
		return nil
	}
	sm := util.NewSignedMessage(s.keyPair, message)
	return sm
}

// processMessagesForever should be run in its own goroutine. This is the only
// thread that is allowed to access the node, because node is not threadsafe.
// The 'unsafe' methods should only be called from within here.
func (s *Server) processMessagesForever() {
	// TODO: run long tests to make sure this is ok
	s.unsafeUpdateOutgoing()

	for {

		select {

		case request := <-s.requests:
			if request.Message != nil {
				response := s.unsafeProcessMessage(request.Message)
				if request.Response != nil {
					request.Response <- response
				}
			}

		case message := <-s.inbox:
			if message != nil {
				s.unsafeProcessMessage(message)
			}

		case <-s.quit:
			break
		}
	}
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.shutdown {
			break
		}
		if err != nil {
			util.Logger.Print("incoming connection error: ", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// Must be called before listen()
// Will retry up to 5 seconds
func (s *Server) acquirePort() {
	s.Logf("listening on port %d", s.port)
	for i := 0; i < 100; i++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
		if err == nil {
			s.listener = ln
			s.start = time.Now()
			return
		}
		time.Sleep(time.Millisecond * time.Duration(50))
	}
	util.Logger.Fatalf("could not acquire port %d", s.port)
}

func (s *Server) broadcast(messages []*util.SignedMessage) {
	for _, message := range messages {
		for _, peer := range s.peers {
			peer.Send(message)
		}
		s.broadcasted += 1
	}
}

// Return a list of everything in a that is not in b.
func subtract(a []*util.SignedMessage, b []*util.SignedMessage) []*util.SignedMessage {
	sigs := make(map[string]bool)
	for _, m := range b {
		sigs[m.Signature()] = true
	}
	answer := []*util.SignedMessage{}
	for _, m := range a {
		if !sigs[m.Signature()] {
			answer = append(answer, m)
		}
	}
	return answer
}

// broadcastIntermittently() sends outgoing messages every so often. It
// should be run as a goroutine. This handles both redundancy rebroadcasts and
// the regular broadcasts of new messages.
func (s *Server) broadcastIntermittently() {
	lastMessages := []*util.SignedMessage{}

	for {
		timer := time.NewTimer(s.RebroadcastInterval)
		select {

		case <-s.quit:
			break

		case messages := <-s.outgoing:
			// See if there are even newer messages
			newerMessages, ok := s.getOutgoing()
			if ok {
				messages = newerMessages
			}

			// When we receive a new outgoing, we only need to send out the
			// messages that have changed since last time.
			changedMessages := subtract(messages, lastMessages)
			lastMessages = messages
			s.broadcast(changedMessages)

		case <-timer.C:
			// It's time for a rebroadcast. Send out duplicate messages.
			// This is a backstop against miscellaneous problems. If the
			// network is functioning perfectly, this isn't necessary.
			s.Logf("performing a backup rebroadcast")
			s.broadcast(lastMessages)
		}
	}
}

func (s *Server) LocalhostAddress() *Address {
	return &Address{
		Host: "127.0.0.1",
		Port: s.port,
	}
}

// ServeForever spawns off all the goroutines and never returns.
// Stop() might not work when you run the server this way, because stopping
// during startup does not work well
func (s *Server) ServeForever() {
	s.acquirePort()

	go s.processMessagesForever()
	go s.listen()
	s.broadcastIntermittently()
}

// ServeInBackground spawns goroutines to run the server.
// It returns once it has successfully bound to its port.
// Stop() should work if it is called after ServeInBackground returns.
func (s *Server) ServeInBackground() {
	s.acquirePort()
	go s.processMessagesForever()
	go s.listen()
	go s.broadcastIntermittently()
}

// ServeHttpInBackground spawns a goroutine to serve the /somethingz urls.
func (s *Server) ServeHttpInBackground(port int) {
	// /healthz just returns OK as long as the server is healthy
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
	})

	http.HandleFunc("/uptimez", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%.2f\n", s.Uptime())
	})

	// /statusz returns more detailed information about this server
	http.HandleFunc("/statusz", func(w http.ResponseWriter, r *http.Request) {
		util.Logger.Print("got /statusz request")
		fmt.Fprintf(w, "%.1fs uptime\n", s.Uptime())
		fmt.Fprintf(w, "%d messages broadcasted\n", s.broadcasted)
		fmt.Fprintf(w, "current slot: %d\n", s.node.Slot())
		fmt.Fprintf(w, "DB_USER: %s\n", os.Getenv("DB_USER"))
		fmt.Fprintf(w, "public key: %s\n", s.keyPair.PublicKey())
		if s.db != nil {
			last := s.db.LastBlock()
			if last == nil {
				fmt.Fprintf(w, "last block: nil\n")
			} else {
				fmt.Fprintf(w, "last block: %s\n", last.String())
			}
			fmt.Fprintf(w, "total database size: %s\n", s.db.TotalSizeInfo())
		}
	})

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	go srv.ListenAndServe()

	go func() {
		<-s.quit
		srv.Shutdown(context.Background())
	}()
}

// Uptime returns uptime in seconds
func (s *Server) Uptime() float64 {
	return time.Now().Sub(s.start).Seconds()
}

func (s *Server) Stats() {
	s.Logf("server stats:")
	s.Logf("%.1fs uptime", s.Uptime())
	s.Logf("%d messages broadcasted", s.broadcasted)
	s.node.Stats()
}

func (s *Server) Stop() {
	s.shutdown = true
	close(s.quit)

	if s.listener != nil {
		s.Logf("releasing port %d", s.port)
		s.listener.Close()
	}

	for _, peer := range s.peers {
		peer.Close()
	}
}
