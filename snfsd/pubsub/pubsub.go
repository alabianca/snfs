package pubsub

import "sync"

type PubSub interface {
	Subscribe(topic string) <-chan []byte
	Publish(topic string, data []byte)
	Close()
}

var psInstance PubSub
var psOnce sync.Once
func getSingleton() PubSub {
	psOnce.Do(func() {
		psInstance = NewPubSub()
	})

	return psInstance
}

func Subscribe(topic string) <-chan []byte {
	return getSingleton().Subscribe(topic)
}

func Publish(topic string, data []byte) {
	getSingleton().Publish(topic, data)
}

func Close() {
	getSingleton().Close()
}

type pubsub struct {
	mtx sync.RWMutex
	subs map[string][]chan []byte
	closed bool
}

func NewPubSub() PubSub {
	return &pubsub{
		mtx:  sync.RWMutex{},
		subs: make(map[string][]chan []byte),
	}
}

func (ps *pubsub) Subscribe(topic string) <-chan []byte {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	ch := make(chan []byte, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)

	return ch
}

func (ps *pubsub) Publish(topic string, data []byte) {
	ps.mtx.RLock()
	defer ps.mtx.RUnlock()

	if ps.closed {
		return
	}

	for _, ch := range ps.subs[topic] {
		dest := make([]byte, len(data))
		copy(dest, data)
		ch <- dest
	}
}

func (ps *pubsub) Close() {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}
