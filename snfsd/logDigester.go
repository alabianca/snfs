package snfsd

import "io"

type LogDigest interface {
	io.Writer
	Process() error
}

type Logger interface {
	Println(v ...interface{})
	Close() error
}
