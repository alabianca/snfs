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
	NodeLookupReq = MessageType(20)
	NodeLookupRes = MessageType(21)
	PingReq       = MessageType(22)
	PingRes       = MessageType(23)
	FindValueReq  = MessageType(24)
	FindValueRes  = MessageType(25)
	StoreReq      = MessageType(26)
	StoreRes      = MessageType(27)
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
	case *NodeLookupRequest:
		*v = NodeLookupRequest{
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
	case NodeLookupReq:
		var nr NodeLookupRequest
		out = &nr
		toKademliaMessage(msg, out)
	case NodeLookupRes:
		var nr NodeLookupResponse
		out = &nr
		toKademliaMessage(msg, out)

	}

	return out
}

func isResponse(msgType MessageType) bool {
	return msgType == NodeLookupRes ||
		msgType == PingRes ||
		msgType == FindValueRes ||
		msgType == StoreRes
}

// Messages

// NODE LOOKUP RESPONSE

type NodeLookupResponse struct {
	senderID     string
	echoRandomID string
	payload      []gokad.Contact
	randomID     string
}

func (n *NodeLookupResponse) MultiplexKey() MessageType {
	return NodeLookupRes
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

func (n *NodeLookupResponse) GetSenderID() string {
	return n.senderID
}

func serializeID(id string) ([]byte, error) {
	return util.BytesFromHex(id)
}




// NODE LOOKUP REQUEST

type NodeLookupRequest struct {
	senderID     string
	echoRandomID string
	payload      string
	randomID     string
}

func newNodeLookupRequest(sID, eID, payload string) *NodeLookupRequest {
	rID := gokad.GenerateRandomID().String()
	if eID == "" {
		eID = rID
	}

	return &NodeLookupRequest{
		senderID: sID,
		echoRandomID: eID,
		payload: payload,
		randomID: rID,
	}
}

func (n *NodeLookupRequest) MultiplexKey() MessageType {
	return NodeLookupReq
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

func (n *NodeLookupRequest) GetSenderID() string {
	return n.senderID
}
