package main

import (
	"encoding/json"

	"github.com/privacylab/talek/libtalek"
)

type talekContactState int

const (
	// A contact is generated with an offer to give to the other party.
	offerGenerated talekContactState = iota
	// A contact has been created from another offer. the initial response is sent.
	answerSent = iota
	// The handshake has been complete.
	confirmed = iota
)

// TalekContact represents a bidirectional message channel with another user.
type TalekContact struct {
	MyNick      string
	TheirName   string
	MyTopic     *libtalek.Topic
	TheirHandle *libtalek.Handle
	state       talekContactState
	outgoing    chan []byte   //msgs from local user to the contact
	incoming    *func([]byte) //msgs from contact to local user
	done        chan byte
}

type talekOffer struct {
	fromNick   string
	readHandle string
	writeTopic string
}

// GetOffer makes a local contact / handle as text for a remote contact.
// the offer contains:
// * A handle for reading messages from me
// * A topic I'll read one msg off of with info on your handle.
func GetOffer(myName, theirName string) (*TalekContact, []byte) {
	contact := new(TalekContact)
	contact.MyNick = myName
	contact.TheirName = theirName
	contact.state = offerGenerated
	var err error
	contact.MyTopic, err = libtalek.NewTopic()
	if err != nil {
		panic(err)
	}
	toPoll, err := libtalek.NewTopic()
	if err != nil {
		panic(err)
	}
	myHandle, err := contact.MyTopic.Handle.MarshalText()
	if err != nil {
		panic(err)
	}

	contact.TheirHandle = &toPoll.Handle
	theirTopic, err := toPoll.MarshalText()
	if err != nil {
		panic(err)
	}

	offer, err := json.Marshal(talekOffer{fromNick: myName, readHandle: string(myHandle), writeTopic: string(theirTopic)})
	if err != nil {
		panic(err)
	}

	return contact, offer
}

// AcceptOffer resolves a remote contact's stream.
func AcceptOffer(myName string, offer []byte, client *libtalek.Client) *TalekContact {
	contact := new(TalekContact)
	contact.MyNick = myName
	contact.state = answerSent
	var err error
	contact.MyTopic, err = libtalek.NewTopic()
	if err != nil {
		panic(err)
	}

	// Deserialize
	offerstruct := talekOffer{}
	if err = json.Unmarshal(offer, &offerstruct); err != nil {
		panic(err)
	}

	rendezvousTopic := libtalek.Topic{}
	if err = rendezvousTopic.UnmarshalText([]byte(offerstruct.writeTopic)); err != nil {
		panic(err)
	}

	contact.TheirHandle = &libtalek.Handle{}
	if err = contact.TheirHandle.UnmarshalText([]byte(offerstruct.readHandle)); err != nil {
		panic(err)
	}

	myHandle, err := contact.MyTopic.Handle.MarshalText()
	if err != nil {
		panic(err)
	}

	if err = client.Publish(&rendezvousTopic, myHandle); err != nil {
		panic(err)
	}
	contact.TheirName = offerstruct.fromNick

	return contact
}

func (t *TalekContact) onMessage(data []byte) bool {
	if t.state == offerGenerated {
		t.TheirHandle = &libtalek.Handle{}
		if err := t.TheirHandle.UnmarshalText(data); err != nil {
			panic(err)
		}
		t.state = confirmed
		return true
	} else if t.state == answerSent {
		t.state = confirmed
		return true
	}
	return false
}

// Channel connects ingoing / outgoing message streams for a running contact.
func (t *TalekContact) Channel(incoming *func([]byte)) chan<- []byte {
	t.incoming = incoming
	return t.outgoing
}

// Start begins polling a contact for messages, and watching for those to send
func (t *TalekContact) Start(c *libtalek.Client) {
	curHandle := t.TheirHandle
	incoming := c.Poll(curHandle)

	t.outgoing = make(chan []byte)
	t.done = make(chan byte, 1)

	go (func() {
		for {
			select {
			case msg := <-incoming:
				if !t.onMessage(msg) && t.incoming != nil {
					(*t.incoming)(msg)
				} else if t.TheirHandle != curHandle {
					c.Done(curHandle)
					curHandle = t.TheirHandle
					incoming = c.Poll(t.TheirHandle)
				}
			case msg := <-t.outgoing:
				c.Publish(t.MyTopic, msg)
			case <-t.done:
				t.outgoing = nil
				return
			}
		}
	})()
}

// Done cleans up a running contact
func (t *TalekContact) Done() {
	if t.outgoing != nil {
		t.done <- 1
	}
}
