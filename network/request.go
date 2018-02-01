package network

import (
	"coinkit/util"
)

// The convention for handling a Request is that you send the response
// to the response channel.
// This happens even if it is a nil response.
type Request struct {
	Message *util.SignedMessage

	// If Message is nil, it has been pre-encoded into MessageString.
	Line string

	Response chan *util.SignedMessage
}

func (r *Request) GetLine() string {
	if r.Message != nil && len(r.Line) == 0 {
		// We need to calculate the message string
		r.Line = util.SignedMessageToLine(r.Message)
	}
	return r.Line
}
