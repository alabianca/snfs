package kadnet

import (
	"errors"

	"github.com/alabianca/gokad"
)

type MessageType int

const NodeLookupType = MessageType(20)

// Errors
const ErrNoMatchMessageSize = "byte length does not match message size"

type Message struct {
	MultiplexKey   MessageType
	SenderID       []byte
	EchoedRandomID []byte
	Payload        []byte
	RandomID       []byte
}

// <-  1 Bytes    <- 20 Bytes  <- 20 Bytes       <- X Bytes  <- 20 Bytes
//  MultiplexKey      SenderID    EchoedRandomId    Payload     RandomID
func process(raw []byte) (Message, error) {
	length := len(raw)
	if length != gokad.MessageSize {
		return Message{}, errors.New(ErrNoMatchMessageSize)
	}

	endOfKey := 1
	endOfSender := 21
	endOfEcho := 41
	endOfRandom := length
	endOfPayload := endOfRandom - 20

	mkey := MessageType(raw[:endOfKey][0])
	senderID := raw[endOfKey:endOfSender]
	echoR := raw[endOfSender:endOfEcho]
	payload := raw[endOfEcho:endOfEcho]
	randomID := raw[endOfPayload:]

	message := Message{
		MultiplexKey:   mkey,
		SenderID:       senderID,
		EchoedRandomID: echoR,
		Payload:        payload,
		RandomID:       randomID,
	}

	return message, nil
}
