package treemap

import "github.com/alabianca/gokad"

type PendingNode interface {
	Contact() gokad.Contact
	Queried() bool
	Answered() bool
}
