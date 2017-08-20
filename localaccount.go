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

		for person := range e.Accounts.Online {
			from.Send([]byte("<presence from='" + person + "@talexmpp/talek' to='" + from.Jid() + "' />"))
		}
	} else if ok {
		// Initiate Contact addition.
		contact, offer := GetOffer()
		contact.Start(e.Client)
		sender := from.Send
		fromUser := contact.Channel(&sender)
		e.Accounts.Online[parsedPresence.To] = fromUser
		from.Send([]byte("<message from='" + parsedPresence.To + "' type='chat'><body>" + string(offer) + "</body></message>"))

	}
}
