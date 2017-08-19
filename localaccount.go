package main

import (
	"fmt"

	"./xmpp"
	"github.com/privacylab/talek/libtalek"

	"sync"
)

// AccountManager deals with understanding the local set of talek logs in use and remote users
type AccountManager struct {
	Online  map[string]chan<- interface{}
	Backend *libtalek.Client
	lock    *sync.Mutex
}

func (a AccountManager) Authenticate(username, password string) (success bool, err error) {
	success = true
	return
}

func (a AccountManager) CreateAccount(username, password string) (success bool, err error) {
	success = true
	return
}

// called periodically by client.
func (a AccountManager) OnlineRoster(jid string) (online []string, err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for person := range a.Online {
		online = append(online, person)
	}
	online = append(online, "Status@talexmpp.local/bot")
	return
}

// msg from local user
func (a AccountManager) routeRoutine(bus <-chan xmpp.Message) {
	var channel chan<- interface{}
	var ok bool

	for {
		message := <-bus
		a.lock.Lock()

		fmt.Printf("%s -> %s\n", message.To, message.Data)
		if channel, ok = a.Online[message.To]; ok {
			channel <- message.Data
		}

		a.lock.Unlock()
	}
}

// Local user online
func (a AccountManager) connectRoutine(bus <-chan xmpp.Connect) {
	for {
		message := <-bus
		a.lock.Lock()
		a.Online[message.Jid] = message.Receiver
		a.lock.Unlock()
	}
}

// Local user offline
func (a AccountManager) disconnectRoutine(bus <-chan xmpp.Disconnect) {
	for {
		message := <-bus
		a.lock.Lock()
		delete(a.Online, message.Jid)
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

type RosterManagementExtension struct {
	Accounts AccountManager
}

func (e *RosterManagementExtension) Process(message interface{}, from *xmpp.Client) {
	parsed, ok := message.(*xmpp.ClientPresence)
	if ok {
		fmt.Printf("Saw client preence: %v\n", parsed)
	}
	if ok && parsed.Type == "subscribe" {
		e.Accounts.Online[parsed.To] = handleMessagesTo(parsed.To)
		fmt.Printf("Saw Subscribe :%v\n", parsed)
	}
	iqparse, ok := message.(*xmpp.ClientIQ)
	if ok && iqparse.Type == "get" {
		fmt.Printf("saw iq req: %s\n", iqparse.Query)
	}
}
