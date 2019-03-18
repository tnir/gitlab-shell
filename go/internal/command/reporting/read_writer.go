package reporting

import "io"

type ReadWriter struct {
	Out    io.Writer
	In     io.Reader
	ErrOut io.Writer
}
