package snfsd

import "io"

type LogDigest interface {
	io.Writer
	Process() error
}
