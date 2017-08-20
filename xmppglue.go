package main

import (
	"fmt"
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
			from.Send([]byte(msg))
			return
		} else if strings.Contains(string(parsed.Query), "jabber:iq:privacy") {
			msg := "<iq id='" + parsed.ID + "' type='result' from='talexmpp' />"
			from.Send([]byte(msg))
			return
		}
	}
	if ok && parsed.Type == "get" && len(parsed.Query) > 0 {
		if strings.TrimSpace(string(parsed.Query)) == `<query xmlns="http://jabber.org/protocol/disco#info"/>` {
			fmt.Printf("Got DiscoInfo query")
			msg := "<iq type='result' from='talexmpp' id='" + parsed.ID + "'>"
			msg = msg + "<query xmlns='http://jabber.org/protocol/disco#info'>"
			msg = msg + "<feature var='http://jabber.org/protocol/disco#info'/>"
			msg = msg + "</query></iq>"
			from.Send([]byte(msg))
			return
		} else if strings.TrimSpace(string(parsed.Query)) == `<query xmlns="jabber:iq:privacy"/>` {
			msg := "<iq type='result' id='" + parsed.ID + "'><query xmlns='jabber:iq:privacy'></query></iq>"
			from.Send([]byte(msg))
			return
		} else if strings.Contains(string(parsed.Query), "jabber:iq:roster") {
			// handled internally
			return
		}
	}
	if ok {
		msg := "<iq type='result' id='" + parsed.ID + "' />"
		from.Send([]byte(msg))
	}
}
