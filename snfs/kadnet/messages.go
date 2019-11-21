package kadnet

import (
	"errors"
	"fmt"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/gokad"
)

type MessageType int

type KademliaMessage interface {
	Serialize() ([]byte, error)
	GetRandomID() string
}

const (
	NodeLookupReq = MessageType(20)
	NodeLookupRes = MessageType(21)
	PingReq       = MessageType(22)
	PingRes       = MessageType(23)
	FindValueReq  = MessageType(24)
	FindValueRes  = MessageType(25)
	StoreReq      = MessageType(26)
	StoreRes      = MessageType(27)
)

func makeMessageChannels(maxBuf int) map[MessageType]chan Message {
	return map[MessageType]chan Message{
		NodeLookupReq: make(chan Message, maxBuf),
		NodeLookupRes: make(chan Message, maxBuf),
		PingReq:       make(chan Message, maxBuf),
		PingRes:       make(chan Message, maxBuf),
		FindValueReq:  make(chan Message, maxBuf),
		FindValueRes:  make(chan Message, maxBuf),
		StoreReq:      make(chan Message, maxBuf),
		StoreRes:      make(chan Message, maxBuf),
	}
}

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

func toKademliaMessage(msg Message, km KademliaMessage) error {
	if km == nil {
		return nil
	}

	switch v := km.(type) {
	case *NodeLookupRequest:
		*v = NodeLookupRequest{
			randomID:     fmt.Sprintf("%x", msg.RandomID),
			echoRandomID: fmt.Sprintf("%x", msg.EchoedRandomID),
			senderID:     fmt.Sprintf("%x", msg.SenderID),
			payload:      fmt.Sprintf("%x", msg.Payload),
		}
	}

	return nil
}

// Messages

type NodeLookupResponse struct {
	senderID     string
	echoRandomID string
	payload      []gokad.Contact
	randomID     string
}

func (n *NodeLookupResponse) Serialize() ([]byte, error) {
	mkey := make([]byte, 1)
	mkey[0] = byte(NodeLookupRes)
	sid, err := serializeID(n.senderID)
	if err != nil {
		return nil, err
	}

	eid, err := serializeID(n.echoRandomID)
	if err != nil {
		return nil, err
	}

	rid, err := serializeID(n.randomID)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0)
	out = append(out, mkey...)
	out = append(out, sid...)
	out = append(out, eid...)
	for _, c := range n.payload {
		ser, err := c.Serialize()
		if err != nil {
			return nil, err
		}

		out = append(out, ser...)
	}

	out = append(out, rid...)

	return out, nil
}

func (n *NodeLookupResponse) GetRandomID() string {
	return n.randomID
}

func serializeID(id string) ([]byte, error) {
	return util.BytesFromHex(id)
}

type NodeLookupRequest struct {
	senderID     string
	echoRandomID string
	payload      string
	randomID     string
}

func (n *NodeLookupRequest) Serialize() ([]byte, error) {
	mkey := make([]byte, 1)
	mkey[0] = byte(NodeLookupReq)

	sid, err := serializeID(n.senderID)
	if err != nil {
		return nil, err
	}

	eid := make([]byte, 20)

	pid, err := serializeID(n.payload)
	if err != nil {
		return nil, nil
	}

	rid, err := serializeID(n.randomID)
	if err != nil {
		return nil, nil
	}

	out := make([]byte, 0)
	out = append(out, sid...)
	out = append(out, eid...)
	out = append(out, pid...)
	out = append(out, rid...)

	return out, nil

}

func (n *NodeLookupRequest) GetRandomID() string {
	return n.randomID
}
