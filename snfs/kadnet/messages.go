package kadnet

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
	message Message
	sender  net.Addr
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

func toKademliaMessage(msg *Message, km KademliaMessage) {

	switch v := km.(type) {
	case *FindNodeRequest:
		*v = FindNodeRequest{
			randomID:     toStringId(msg.RandomID),
			echoRandomID: toStringId(msg.EchoedRandomID),
			senderID:     toStringId(msg.SenderID),
			payload:      toStringId(msg.Payload),
		}
	case *FindNodeResponse:
		*v = FindNodeResponse{
			senderID:     toStringId(msg.SenderID),
			echoRandomID: toStringId(msg.EchoedRandomID),
			payload:      processContacts(msg.Payload),
			randomID:     toStringId(msg.RandomID),
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
	id, err := gokad.From(toStringId(idBytes))
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

func processMessage(msg *Message) KademliaMessage {
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

func (n *FindNodeResponse) Bytes() ([]byte, error) {
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

func toStringId(id []byte) string {
	return fmt.Sprintf("%x", id)
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
		senderID:     sID,
		echoRandomID: eID,
		payload:      payload,
		randomID:     rID,
	}
}

func (n *FindNodeRequest) MultiplexKey() MessageType {
	return FindNodeReq
}

func (n *FindNodeRequest) Bytes() ([]byte, error) {
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
