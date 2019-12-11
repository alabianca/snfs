package messages

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
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
	PingReqSize = 41
	PingResSize = 41
	PingReqResSize = 61
	FindNodeReqSize = 61
	FindValueReqSize = 61
	StoreReqSize = 67
	FindNodeResSize = 581
	FindValueResSize = 201 // Note: Assumes there are always k values in the payload
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

// <-  1 Bytes    <- 20 Bytes  <- 20 Bytes       <- X Bytes  <- 20 Bytes
//  MultiplexKey      SenderID    EchoedRandomId    Payload     RandomID
func Process(raw []byte) (Message, error) {
	var message Message
	mKey := MessageType(raw[0])
	log.Printf("Processing %d %d\n", mKey, len(raw))
	switch mKey {
	case FindNodeReq:
		processFindNodeRequest(&message, raw[1:])
	}

	return message, nil
}

func processFindNodeRequest(m *Message, p []byte) error {
	// account for the multiplex key not to be there at this point
	if len(p) != FindNodeReqSize - 1{
		return errors.New(ErrNoMatchMessageSize)
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


