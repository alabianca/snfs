package messages

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/gokad"
)

type MessageType int

type KademliaMessage interface {
	MultiplexKey() MessageType
	Bytes() ([]byte, error)
	GetRandomID() string
	GetSenderID() string
}

const (
	NodeLookup   = MessageType(30)
	FindNodeReq  = MessageType(20)
	FindNodeRes  = MessageType(21)
	PingReq      = MessageType(22)
	PingRes      = MessageType(23)
	FindValueReq = MessageType(24)
	FindValueRes = MessageType(25)
	StoreReq     = MessageType(26)
	StoreRes     = MessageType(27)
)

// Errors
const ErrNoMatchMessageSize = "byte length does not match message size"

type CompleteMessage struct {
	Message Message
	Sender  net.Addr
}

type Message struct {
	MultiplexKey   MessageType
	SenderID       []byte
	EchoedRandomID []byte
	Payload        []byte
	RandomID       []byte
}

// <-  1 Bytes    <- 20 Bytes  <- 20 Bytes       <- X Bytes  <- 20 Bytes
//  MultiplexKey      SenderID    EchoedRandomId    Payload     RandomID
func Process(raw []byte) (Message, error) {
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

func ToKademliaMessage(msg *Message, km KademliaMessage) {

	switch v := km.(type) {
	case *FindNodeRequest:
		*v = FindNodeRequest{
			RandomID:     ToStringId(msg.RandomID),
			EchoRandomID: ToStringId(msg.EchoedRandomID),
			SenderID:     ToStringId(msg.SenderID),
			Payload:      ToStringId(msg.Payload),
		}
	case *FindNodeResponse:
		*v = FindNodeResponse{
			SenderID:     ToStringId(msg.SenderID),
			EchoRandomID: ToStringId(msg.EchoedRandomID),
			Payload:      processContacts(msg.Payload),
			RandomID:     ToStringId(msg.RandomID),
		}
	}

}

func processContacts(raw []byte) []gokad.Contact {

	split := bytes.Split(raw, []byte("/"))
	out := make([]gokad.Contact, len(split))

	insert := 0
	for _, c := range split {
		contact, err := toContact(c)
		if err == nil {
			out[insert] = contact
			insert++
		}
	}

	return out
}

func toContact(b []byte) (gokad.Contact, error) {
	l := len(b)
	if l == 0 {
		return gokad.Contact{}, errors.New("Empty Contact")
	}

	idOffset := 0
	portOffset := 20
	ipOffset := 22

	if l <= ipOffset {
		return gokad.Contact{}, errors.New("Malformed Contact")
	}

	idBytes := b[idOffset:portOffset]
	portBytes := b[portOffset:ipOffset]
	ipBytes := b[ipOffset:len(b)]
	ip := net.ParseIP(string(ipBytes))

	if ip == nil {
		return gokad.Contact{}, errors.New("Invalid IP")
	}

	port := binary.BigEndian.Uint16(portBytes)
	id, err := gokad.From(ToStringId(idBytes))
	if err != nil {
		return gokad.Contact{}, errors.New("Invalid ID")
	}

	c := gokad.Contact{
		ID:   id,
		IP:   ip,
		Port: int(port),
	}

	return c, nil

}

func ProcessMessage(msg *Message) KademliaMessage {
	var out KademliaMessage
	switch msg.MultiplexKey {
	case FindNodeReq:
		var nr FindNodeRequest
		out = &nr
		ToKademliaMessage(msg, out)
	case FindNodeRes:
		var nr FindNodeResponse
		out = &nr
		ToKademliaMessage(msg, out)

	}

	return out
}

func IsResponse(msgType MessageType) bool {
	return msgType == FindNodeRes ||
		msgType == PingRes ||
		msgType == FindValueRes ||
		msgType == StoreRes
}

func SerializeID(id string) ([]byte, error) {
	return util.BytesFromHex(id)
}

func ToStringId(id []byte) string {
	return fmt.Sprintf("%x", id)
}


