package kadnet

import (
	"errors"
	"fmt"
	"net"

	"github.com/alabianca/snfs/util"

	"github.com/alabianca/gokad"
)

type MessageType int

type KademliaMessage interface {
	MultiplexKey() MessageType
	Serialize() ([]byte, error)
	GetRandomID() string
	GetSenderID() string
}

const (
	FindNodeReq  = MessageType(20)
	FindNodeRes  = MessageType(21)
	PingReq      = MessageType(22)
	PingRes      = MessageType(23)
	FindValueReq = MessageType(24)
	FindValueRes = MessageType(25)
	StoreReq     = MessageType(26)
	StoreRes     = MessageType(27)
)

func makeMessageChannels(maxBuf int, messageTypes ...MessageType) map[MessageType]chan CompleteMessage {
	m := make(map[MessageType]chan CompleteMessage)

	for _, t := range messageTypes {
		m[t] = make(chan CompleteMessage, maxBuf)
	}

	return m
}

// Errors
const ErrNoMatchMessageSize = "byte length does not match message size"

type CompleteMessage struct {
	message KademliaMessage
	sender  *net.UDPAddr
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

func toKademliaMessage(msg Message, km KademliaMessage) {

	switch v := km.(type) {
	case *FindNodeRequest:
		*v = FindNodeRequest{
			randomID:     fmt.Sprintf("%x", msg.RandomID),
			echoRandomID: fmt.Sprintf("%x", msg.EchoedRandomID),
			senderID:     fmt.Sprintf("%x", msg.SenderID),
			payload:      fmt.Sprintf("%x", msg.Payload),
		}
	}

}

func processMessage(msg Message) KademliaMessage {
	var out KademliaMessage
	switch msg.MultiplexKey {
	case FindNodeReq:
		var nr FindNodeRequest
		out = &nr
		toKademliaMessage(msg, out)
	case FindNodeRes:
		var nr FindNodeResponse
		out = &nr
		toKademliaMessage(msg, out)

	}

	return out
}

func isResponse(msgType MessageType) bool {
	return msgType == FindNodeRes ||
		msgType == PingRes ||
		msgType == FindValueRes ||
		msgType == StoreRes
}

// Messages

// FIND NODE RESPONSE

type FindNodeResponse struct {
	senderID     string
	echoRandomID string
	payload      []gokad.Contact
	randomID     string
}

func (n *FindNodeResponse) MultiplexKey() MessageType {
	return FindNodeRes
}

func (n *FindNodeResponse) Serialize() ([]byte, error) {
	mkey := make([]byte, 1)
	mkey[0] = byte(FindNodeRes)
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

func (n *FindNodeResponse) GetRandomID() string {
	return n.randomID
}

func (n *FindNodeResponse) GetSenderID() string {
	return n.senderID
}

func serializeID(id string) ([]byte, error) {
	return util.BytesFromHex(id)
}




// FIND NODE REQUEST

type FindNodeRequest struct {
	senderID     string
	echoRandomID string
	payload      string
	randomID     string
}

func newFindNodeRequest(sID, eID, payload string) *FindNodeRequest {
	rID := gokad.GenerateRandomID().String()
	if eID == "" {
		eID = rID
	}

	return &FindNodeRequest{
		senderID: sID,
		echoRandomID: eID,
		payload: payload,
		randomID: rID,
	}
}

func (n *FindNodeRequest) MultiplexKey() MessageType {
	return FindNodeReq
}

func (n *FindNodeRequest) Serialize() ([]byte, error) {
	mkey := make([]byte, 1)
	mkey[0] = byte(FindNodeReq)

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

func (n *FindNodeRequest) GetRandomID() string {
	return n.randomID
}

func (n *FindNodeRequest) GetSenderID() string {
	return n.senderID
}
