package protocol

import (
	"bufio"
	"errors"
	"net"
	"time"

	"github.com/nnsgmsone/units/breaker"
)

var (
	TYPEERROR   = errors.New("Unsupport Type")
	ENCODEERROR = errors.New("Failed To Encode")
)

type Unit interface {
	Remove() error
	Recv() (*Message, error)
	Send(string, interface{}) error
	SendAndRecv(string, interface{}) (*Message, error)
}

type Server interface {
	Run()
	Stop()
}

type Message struct {
	Name string
	Msg  interface{}
}

type MessageError struct {
	M string
}

type MessageSlice struct {
	M string
}

type MessageInteger struct {
	M int64
}

type MessageArray struct {
	M []string
}

type MessageWriter interface {
	Write(string, interface{}) error
	WriteMessage(*Message) error
}

type DealFunc (func(interface{}, MessageWriter, *Message))

type connection struct {
	state int
	s     *server
	c     net.Conn
}

type messageWriter struct {
	w *bufio.Writer
}

type server struct {
	df  DealFunc
	usr interface{}
	lis net.Listener
	ch  chan struct{}
	brk breaker.Breaker
}

type unit struct {
	rc      int // retry limit
	addr    string
	conn    net.Conn
	timeout time.Duration
}
