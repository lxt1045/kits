package gob

import (
	"errors"
)

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var ErrShutdown = errors.New("connection is shut down")
var ErrNotSuportType = errors.New("type is not suported")
var ErrPendingTooLong = errors.New("call is pending too long")
var ErrSendError = errors.New("got error when send to network")

type Header struct {
	MsgNO uint32  //
	next  *Header // for free list in Server
}

type Codec interface {
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	// Write must be safe for concurrent use by multiple goroutines.
	Write(header *Header, body interface{}) error

	Close() error
}
