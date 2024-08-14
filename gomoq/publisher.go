package gomoq

import (
	"errors"
)

type Publisher struct {
	/*
	 * Client
	 * If this is not initialized, use default Client
	 */
	Client
	// Bidirectional stream to send controll message

	// Unidirectional stream to send object message

	// A map of our announced tracks

}

func (p Publisher) Connect(url string) error {
	// Check if the Client specify the Versions
	if len(p.clientSetupMessage.Versions) < 1 {
		return errors.New("no versions is specifyed")
	}

	// Add role parameter as publisher
	p.clientSetupMessage.Parameters.addIntParameter(role, uint64(pub))

	// Connect to the server
	err := p.connect(url)
	if err != nil {
		return err
	}

	return nil
}

func (p Publisher) ReceiveMessage() {
	// handle each message
}
func (Publisher) receiveAnnounceOk() {
	// turn state of announce to "ack"
}
func (Publisher) receiveAnnounceError() {
	// turn state of announce to the error
}
func (Publisher) receiveSubscribe() {
	// send SUBSCRIBE_OK
}
func (Publisher) receiveUnsubscribe() {}
func (Publisher) SendMessages() {
	// handle each message
}

func (p Publisher) Announce(a AnnounceMessage) {
	// send ANNOUNCE
	// Send SETUP_CLIENT message
	p.controlStream.Write(a.serialize())
}
func (Publisher) Subscribed() {
	// send SUBSCRIBE_OK
}
