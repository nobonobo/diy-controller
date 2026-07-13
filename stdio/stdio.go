package stdio

import (
	"errors"
	"io"
	"os"

	"github.com/nobonobo/diy-controller/board"
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

func (rw *ReadWriteCloser) Write(p []byte) (int, error) {
	board.LED2.Low()
	defer board.LED2.High()
	return rw.Writer.Write(p)
}

func (rw *ReadWriteCloser) Read(p []byte) (int, error) {
	board.LED1.Low()
	defer board.LED1.High()
	return rw.Reader.Read(p)
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
