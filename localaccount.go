package main

import (
	"fmt"
	"strings"

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
		localPart := strings.SplitN(message.Jid, "@", 1)
		a.Online[localPart[0]] = message.Receiver
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

// RosterManagementExtension watches for messages outbound from client.
type RosterManagementExtension struct {
	Accounts AccountManager
}

// Process takes in messages
func (e *RosterManagementExtension) Process(message interface{}, from *xmpp.Client) {
	parsed, ok := message.(*xmpp.ClientPresence)
	if ok {
		fmt.Printf("Saw client presence: %v\n", parsed)
	}
}
