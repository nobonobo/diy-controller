package stdio

import (
	"errors"
	"io"
	"os"
)

type ReadWriteCloser struct {
	io.Reader
	io.Writer

	rCloser io.Closer
	wCloser io.Closer
}

func New(r io.ReadCloser, w io.WriteCloser) io.ReadWriteCloser {
	return &ReadWriteCloser{
		Reader:  r,
		Writer:  w,
		rCloser: r,
		wCloser: w,
	}
}

func (rw *ReadWriteCloser) Close() error {
	werr := rw.wCloser.Close()
	rerr := rw.rCloser.Close()
	if werr != nil && rerr != nil {
		return errors.Join(werr, rerr)
	}
	if werr != nil {
		return werr
	}
	return rerr
}

func NewStdio() io.ReadWriteCloser {
	return New(os.Stdin, os.Stdout)
}
