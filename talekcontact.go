package main

import "github.com/privacylab/talek/libtalek"

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
	MyTopic     *libtalek.Topic
	TheirHandle *libtalek.Handle
	state       talekContactState
}

// GetOffer makes a local contact / handle as text for a remote contact.
// the offer contains:
// * A handle for reading messages from me
// * A topic I'll read one msg off of with info on your handle.
func GetOffer() (*TalekContact, []byte) {
	contact := new(TalekContact)
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

	return contact, append(theirTopic, myHandle...)
}

// AcceptOffer resolves a remote contact's stream.
func AcceptOffer(offer []byte, client *libtalek.Client) *TalekContact {
	contact := new(TalekContact)
	contact.state = answerSent
	var err error
	contact.MyTopic, err = libtalek.NewTopic()
	if err != nil {
		panic(err)
	}

	// Learn how long a serialized topic is:
	rendezvous, err := contact.MyTopic.MarshalText()
	if err != nil {
		panic(err)
	}
	topicLen := len(rendezvous)

	// Deserialize
	rendezvousTopic := libtalek.Topic{}
	if err = rendezvousTopic.UnmarshalText(offer[0:topicLen]); err != nil {
		panic(err)
	}

	contact.TheirHandle = &libtalek.Handle{}
	if err = contact.TheirHandle.UnmarshalText(offer[topicLen:]); err != nil {
		panic(err)
	}

	myHandle, err := contact.MyTopic.Handle.MarshalText()
	if err != nil {
		panic(err)
	}

	if err = client.Publish(&rendezvousTopic, myHandle); err != nil {
		panic(err)
	}

	return contact
}
