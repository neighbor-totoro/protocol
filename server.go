package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/nnsgmsone/units/breaker"
)

func New(port int, usr interface{}, brk breaker.Breaker, df DealFunc) *server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil
	}
	return &server{df, usr, lis, make(chan struct{}), brk}
}

func (s *server) Run() {
	ch := make(chan net.Conn)
	go func() {
		for {
			conn, err := s.lis.Accept()
			if err != nil {
				for {
					lis, err := net.Listen("tcp", s.lis.Addr().String())
					if err == nil {
						s.lis.Close()
						s.lis = lis
						break
					}
				}
			}
			ch <- conn
		}
	}()
	for {
		select {
		case <-s.ch:
			s.ch <- struct{}{}
			return
		case conn := <-ch:
			go s.brk.NewConnection(&connection{0, s, conn})
		}
	}
}

func (s *server) Stop() {
	s.ch <- struct{}{}
	<-s.ch
	close(s.ch)
	s.lis.Close()
}

func (c *connection) Response() error {
	mw := NewMessageWriter(bufio.NewWriter(c.c))
	msg, err := ReadMessage(bufio.NewReader(c.c))
	if err != nil {
		c.state = -1
		if err != io.EOF {
			mw.Write("", errors.New("illegal request"))
		}
		return err
	}
	c.s.df(c.s.usr, mw, msg)
	return nil
}

func (c *connection) Close() error {
	return c.c.Close()
}

func (c *connection) Serve() error {
	for {
		if err := c.s.brk.NewRequest(c); err != nil {
			NewMessageWriter(bufio.NewWriter(c.c)).Write("", err)
			return err
		}
		if c.state != 0 {
			return errors.New("break")
		}
	}
	return nil
}
