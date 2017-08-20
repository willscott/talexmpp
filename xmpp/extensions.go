package xmpp

import (
	"fmt"
	"log"
	"strings"
)

// Extension interface for processing normal messages
type Extension interface {
	Process(message interface{}, from *Client)
}

// DebugExtension just dumps data
type DebugExtension struct {
}

// Process a message (write to debug logger)
func (e *DebugExtension) Process(message interface{}, from *Client) {
	log.Printf("Processing message: %s", message)
}

// NormalMessageExtension handles client messages
type NormalMessageExtension struct {
	MessageBus chan<- Message
}

// Process sends `ClientMessage`s from a client down the `MessageBus`
func (e *NormalMessageExtension) Process(message interface{}, from *Client) {
	parsed, ok := message.(*ClientMessage)
	if ok {
		e.MessageBus <- Message{To: parsed.To, Data: message}
	}
}

// RosterExtension handles ClientIQ presence requests and updates
type RosterExtension struct {
	Accounts AccountManager
}

// Process responds to Presence requests from a client
func (e *RosterExtension) Process(message interface{}, from *Client) {
	parsed, ok := message.(*ClientIQ)

	// handle things we need to handle
	//TODO: actually parse queries to support less brittle xml matching
	if ok && (strings.TrimSpace(string(parsed.Query)) == "<query xmlns='jabber:iq:roster'/>" ||
		strings.TrimSpace(string(parsed.Query)) == `<query xmlns="jabber:iq:roster"/>` ||
		string(parsed.Query) == "<query xmlns='jabber:iq:roster' ></query>") {
		// respond with roster
		roster, _ := e.Accounts.OnlineRoster(from.jid)
		msg := "<iq id='" + parsed.ID + "' to='" + from.jid + "' type='result'><query xmlns='jabber:iq:roster' ver='ver7'>"
		for _, v := range roster {
			msg = msg + "<item jid='" + v + "@talexmpp' name='" + v + "'/>"
		}
		msg = msg + "</query></iq>"

		fmt.Printf("sending back roster.\n")
		// respond to client
		from.messages <- []byte(msg)
	}
}
