package snfsd

type Publisher interface {
	Publish(topic string, data []byte)
}

type Subscriber interface {
	Subscribe(topic string) <-chan []byte
}

type PubSub interface {
	Publisher
	Subscriber
	Close()
}
