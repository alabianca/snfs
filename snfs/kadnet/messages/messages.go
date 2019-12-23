package messages

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
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
	// Message Types
	NodeLookup   = MessageType(30)
	FindNodeReq  = MessageType(20)
	FindNodeRes  = MessageType(21)
	PingReq      = MessageType(22)
	PingRes      = MessageType(23)
	FindValueReq = MessageType(24)
	FindValueRes = MessageType(25)
	StoreReq     = MessageType(26)
	StoreRes     = MessageType(27)
	// Message Sizes
	PingReqSize      = 41
	PingResSize      = 41
	PingReqResSize   = 61
	FindNodeReqSize  = 61
	FindValueReqSize = 61
	StoreReqSize     = 79
	FindNodeResSize  = 841
	FindValueResSize = 441 // Note: Assumes there are always k values in the payload
)

// Errors
const ErrNoMatchMessageSize = "byte length does not match message size"
const ErrContactsMalformed = "contacts malformed. should be an integer"

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

func GetMessageSize(x MessageType) int {
	var size int
	switch x {
	case FindNodeReq:
		size = FindNodeReqSize
	case FindNodeRes:
		size = FindNodeResSize
	case FindValueReq:
		size = FindValueReqSize
	case FindValueRes:
		size = FindValueResSize
	case PingReq:
		size = PingReqSize
	case PingRes:
		size = PingResSize
	case StoreReq:
		size = StoreReqSize

	}

	return size
}

// Process - General structure of a message
// <-  1 Bytes    <- 20 Bytes  <- 20 Bytes       <- X Bytes  <- 20 Bytes
//  MultiplexKey      SenderID    EchoedRandomId    Payload     RandomID
func Process(raw []byte) (Message, error) {
	var message Message
	mKey := MessageType(raw[0])
	switch mKey {
	case FindNodeReq:
		processFindNodeRequest(&message, raw[1:])
	case FindNodeRes:
		processFindNodeResponse(&message, raw[1:])

	}

	return message, nil
}

func processFindNodeRequest(m *Message, p []byte) error {
	if err := checkBounds(p, FindNodeReqSize); err != nil {
		return err
	}
	offsetSender := 0
	offsetLookupId := 20
	offsetRandomId := 40

	m.MultiplexKey = FindNodeReq
	m.SenderID = p[offsetSender:offsetLookupId]
	m.Payload = p[offsetLookupId:offsetRandomId]
	m.RandomID = p[offsetRandomId:len(p)]

	return nil

}

func processFindNodeResponse(m *Message, p []byte) error {
	l := len(p)

	offsetSender := 0
	offsetEchoRandomId := 20
	offsetPayload := 40
	offsetRandomId := len(p) - 20

	// ensure there is space for senderId and echo random id
	if l < offsetPayload - 1 {
		return errors.New(ErrNoMatchMessageSize)
	}

	// good. now ensure there is space for at least one contact and a randomId (every contact is 38 bytes long)
	contactL := 38
	if l < (offsetPayload + contactL + 20) {
		return errors.New(ErrNoMatchMessageSize)
	}


	m.MultiplexKey = FindNodeRes
	m.SenderID = p[offsetSender:offsetEchoRandomId]
	m.EchoedRandomID = p[offsetEchoRandomId:offsetPayload]
	m.Payload = p[offsetPayload: offsetRandomId]
	m.RandomID = p[offsetRandomId:]

	return nil
}

func checkBounds(p []byte, size int) error {
	// account for the multiplex key not to be there at this point
	if len(p) != size -1 {
		return errors.New(ErrNoMatchMessageSize)
	}

	return nil
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
		if c, err := processContacts(msg.Payload); err == nil {
			*v = FindNodeResponse{
				SenderID:     ToStringId(msg.SenderID),
				EchoRandomID: ToStringId(msg.EchoedRandomID),
				Payload:      c,
				RandomID:     ToStringId(msg.RandomID),
			}
		}

	}

}

// every contact is 38 bytes long.
// split the raw bytes in chuncks of 38 bytes.
func processContacts(raw []byte) ([]gokad.Contact, error) {
	offset := 0
	cLen := 38
	l := float64(len(raw))
	x := float64(cLen)
	numContacts := l / x
	// if numContacts if not a flat integer we have some malformed payload.
	if numContacts != math.Trunc(numContacts) {
		return nil, errors.New(ErrContactsMalformed)
	}

	out := make([]gokad.Contact, int(numContacts))

	insert := 0
	for offset < int(l) {
		contact, err := toContact(raw[offset:offset+cLen])
		if err == nil {
			out[insert] = contact
			insert++
		}
		offset += cLen
	}

	return out, nil
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
	ip := net.IP(ipBytes)


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

func IsValid(msgType MessageType) bool {
	return IsRequest(msgType) || IsResponse(msgType)
}

func IsResponse(msgType MessageType) bool {
	return msgType == FindNodeRes ||
		msgType == PingRes ||
		msgType == FindValueRes ||
		msgType == StoreRes
}

func IsRequest(msgType MessageType) bool {
	return msgType == FindNodeReq ||
		msgType == PingReq ||
		msgType == FindValueReq ||
		msgType == StoreReq
}

func SerializeID(id string) ([]byte, error) {
	return util.BytesFromHex(id)
}

func ToStringId(id []byte) string {
	return fmt.Sprintf("%x", id)
}
