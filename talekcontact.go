package main

import "github.com/privacylab/talek/libtalek"

// TalekContact represents a bidirectional message channel with another user.
type TalekContact struct {
	MyTopic     *libtalek.Topic
	TheirHandle *libtalek.Handle
}

// GetOffer makes a local contact / handle as text for a remote contact.
func GetOffer() (*TalekContact, []byte) {
	contact := new(TalekContact)
	var err error
	contact.MyTopic, err = libtalek.NewTopic()
	if err != nil {
		panic(err)
	}
	offer, err := contact.MyTopic.Handle.MarshalText()
	if err != nil {
		panic(err)
	}
	return contact, offer
}

// CompleteOffer resolves a remote contact's stream.
func (t *TalekContact) CompleteOffer(other []byte) bool {
	if t.TheirHandle != nil {
		return false
	}
	var err error
	t.TheirHandle, err = libtalek.NewHandle()
	if err != nil {
		t.TheirHandle = nil
		return false
	}
	err = t.TheirHandle.UnmarshalText(other)
	if err != nil {
		t.TheirHandle = nil
		return false
	}
	return true
}
