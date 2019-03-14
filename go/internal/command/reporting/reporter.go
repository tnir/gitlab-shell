package reporting

import "io"

type Reporter struct {
	Out    io.Writer
	In     io.Reader
	ErrOut io.Writer
}
