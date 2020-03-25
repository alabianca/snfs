package snfsd

type Watchdog interface {
	Watch()
	Close()
}
