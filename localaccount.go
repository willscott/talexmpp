package main

import (
	"encoding/xml"
	"fmt"
	"strings"

	"./xmpp"
	"github.com/privacylab/talek/libtalek"

	"sync"
)

// AccountManager deals with understanding the local set of talek logs in use and remote users
type AccountManager struct {
	Online  map[string]chan<- []byte
	Backend *libtalek.Client
	lock    *sync.Mutex
}

// Authenticate called when local user attempts to authenticate
func (a AccountManager) Authenticate(username, password string) (success bool, err error) {
	success = true
	return
}

// CreateAccount called when local iser attempts to register
func (a AccountManager) CreateAccount(username, password string) (success bool, err error) {
	success = true
	return
}

// OnlineRoster is called periodically by client
func (a AccountManager) OnlineRoster(jid string) (online []string, err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	// For status
	online = append(online, "status")
	for person := range a.Online {
		online = append(online, person)
	}
	return
}

// msg from local user
func (a AccountManager) routeRoutine(bus <-chan xmpp.Message) {
	var channel chan<- []byte
	var ok bool

	for {
		message := <-bus
		a.lock.Lock()

		fmt.Printf("%s -> %s\n", message.To, message.Data)
		if channel, ok = a.Online[message.To]; ok {
			switch message.Data.(type) {
			case []byte:
				channel <- message.Data.([]byte)
			default:
				data, err := xml.Marshal(message.Data)
				if err != nil {
					panic(err)
				}
				channel <- data
			}
		}

		a.lock.Unlock()
	}
}

// Local user online
func (a AccountManager) connectRoutine(bus <-chan xmpp.Connect) {
	for {
		message := <-bus
		a.lock.Lock()
		localPart := strings.SplitN(message.Jid, "@", 2)
		fmt.Printf("Adding %s to roster\n", localPart)
		a.Online[localPart[0]] = message.Receiver
		a.lock.Unlock()
	}
}

// Local user offline
func (a AccountManager) disconnectRoutine(bus <-chan xmpp.Disconnect) {
	for {
		message := <-bus
		a.lock.Lock()
		localPart := strings.SplitN(message.Jid, "@", 2)
		delete(a.Online, localPart[0])
		a.lock.Unlock()
	}
}

func handleMessagesTo(jid string) chan interface{} {
	iface := make(chan interface{})
	go func() {
		for {
			n := <-iface
			fmt.Printf("Msg To %s: %s\n", jid, n)
		}
	}()
	return iface
}

// RosterManagementExtension watches for messages outbound from client.
type RosterManagementExtension struct {
	Accounts AccountManager
	Client   *libtalek.Client
}

// Process takes in messages
func (e *RosterManagementExtension) Process(message interface{}, from *xmpp.Client) {
	parsedPresence, ok := message.(*xmpp.ClientPresence)
	if ok && parsedPresence.Type != "subscribe" {
		fmt.Printf("Saw client presence: %v\n", parsedPresence)
		from.Send([]byte("<presence from='status@talexmpp'><priority>1</priority></presence>"))
		for person := range e.Accounts.Online {
			from.Send([]byte("<presence from='" + person + "@talexmpp/talek' to='" + from.Jid() + "' />"))
		}
	} else if ok {
		// Initiate Contact addition.
		//TODO: parse <nick> from subscription request
		contact, offer := GetOffer("nickname", parsedPresence.To)
		contact.Start(e.Client)
		sender := func(msg []byte) {
			wrapped := fmt.Sprintf("<message from='%s@talexmpp/talek' type='chat'><body>%s</body></message>", parsedPresence.To, msg)
			from.Send([]byte(wrapped))
		}
		fromUser := contact.Channel(&sender)
		e.Accounts.Online[parsedPresence.To] = fromUser
		from.Send([]byte("<message from='status@talexmpp' type='chat'><body>Contact generated. Offer:\n" + string(offer) + "</body></message>"))
	}

	parsedMessage, ok := message.(*xmpp.ClientMessage)
	if ok && parsedMessage.To == "status@talexmpp" {
		// accept contact offer.
		offer := parsedMessage.Body
		// TODO: parse name from config or local contact
		contact := AcceptOffer("nickname", []byte(offer), e.Client)
		if contact == nil {
			from.Send([]byte("<message from='status@talexmpp' type='chat'><body>Failed to accept offer.</body></message>"))
			return
		}
		contact.Start(e.Client)
		sender := func(msg []byte) {
			wrapped := fmt.Sprintf("<message from='%s@talexmpp/talek' type='chat'><body>%s</body></message>", contact.TheirName, msg)
			from.Send([]byte(wrapped))
		}
		fromUser := contact.Channel(&sender)
		e.Accounts.Online[contact.TheirName] = fromUser
	}
}
