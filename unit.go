package protocol

import (
	"bufio"
	"net"
	"time"
)

func NewUnit(rc int, addr string, timeout time.Duration) *unit {
	u := new(unit)
	u.rc = rc
	u.addr = addr
	u.timeout = timeout
	u.conn, _ = net.Dial("tcp", u.addr)
	return u
}

func (u *unit) Remove() error {
	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
	}
	return nil
}

func (u *unit) Recv() (*Message, error) {
	u.conn.SetReadDeadline(time.Now().Add(u.timeout))
	return ReadMessage(bufio.NewReader(u.conn))
}

func (u *unit) Send(name string, msg interface{}) error {
	var err error

	if err = u.reConnect(); err != nil {
		return err
	}
	u.conn.SetWriteDeadline(time.Now().Add(u.timeout))
	switch err = NewMessageWriter(bufio.NewWriter(u.conn)).Write(name, msg); err {
	case nil:
		return nil
	case TYPEERROR, ENCODEERROR:
		return err
	default: // retry
		for i := 0; i < u.rc; i++ {
			if err = u.connect(); err != nil {
				continue
			}
			u.conn.SetWriteDeadline(time.Now().Add(u.timeout))
			if err = NewMessageWriter(bufio.NewWriter(u.conn)).Write(name, msg); err != nil {
				continue
			}
			return nil
		}
		u.Remove()
		return err
	}
}

func (u *unit) SendAndRecv(name string, msg interface{}) (*Message, error) {
	if err := u.Send(name, msg); err != nil {
		return nil, err
	}
	return u.Recv()
}

func (u *unit) connect() error {
	var err error

	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
	}
	if u.conn, err = net.Dial("tcp", u.addr); err != nil {
		return err
	}
	return nil
}

func (u *unit) reConnect() error {
	if u.conn == nil {
		return u.connect()
	}
	return nil
}
