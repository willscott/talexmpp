package main

import (
	"strings"

	"./xmpp"
)

// GlueExtension finishes session establishment by handling the iq set session stanza.
type GlueExtension struct {
}

// Process messages
func (e *GlueExtension) Process(message interface{}, from *xmpp.Client) {
	parsed, ok := message.(*xmpp.ClientIQ)

	if ok && parsed.Type == "set" && len(parsed.Query) > 0 {
		if strings.TrimSpace(string(parsed.Query)) == `<session xmlns="urn:ietf:params:xml:ns:xmpp-session"/>` {
			msg := "<iq id='" + parsed.ID + "' type='result' from='talexmpp' />"
			from.Send(msg)
		}
	}
}
