package protocol

import (
	"bufio"
	"encoding/gob"
	"errors"

	"github.com/infinivision/common/miscellaneous"
)

func init() {
	gob.Register(MessageError{})
	gob.Register(MessageArray{})
}

func NewMessageWriter(w *bufio.Writer) *messageWriter {
	return &messageWriter{w}
}

func NewMessage(name string, x interface{}) *Message {
	switch v := x.(type) {
	case error:
		return &Message{name, MessageError{v.Error()}}
	case []string:
		return &Message{name, MessageArray{v}}
	default:
		return nil
	}
}

func (m *messageWriter) Write(name string, x interface{}) error {
	var msg *Message

	switch v := x.(type) {
	case error:
		msg = &Message{name, MessageError{v.Error()}}
	case []string:
		msg = &Message{name, MessageArray{v}}
	default:
		return TYPEERROR
	}
	data, err := miscellaneous.Encode(msg)
	if err != nil {
		return ENCODEERROR
	}
	length := miscellaneous.E64func(uint64(len(data)))
	if _, err := m.w.Write(length); err != nil {
		return err
	}
	if _, err := m.w.Write(data); err != nil {
		return err
	}
	if err := m.w.Flush(); err != nil {
		return err
	}
	return nil
}

func (m *messageWriter) WriteMessage(msg *Message) error {
	data, err := miscellaneous.Encode(msg)
	if err != nil {
		return ENCODEERROR
	}
	length := miscellaneous.E64func(uint64(len(data)))
	if _, err := m.w.Write(length); err != nil {
		return err
	}
	if _, err := m.w.Write(data); err != nil {
		return err
	}
	return m.w.Flush()

}

func ReadMessage(r *bufio.Reader) (*Message, error) {
	var msg Message

	buf := make([]byte, 8)
	if n, err := r.Read(buf); n != 8 {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("Illegal Length")
	}
	length, err := miscellaneous.D64func(buf)
	if err != nil {
		return nil, err
	}
	buf = make([]byte, length)
	if n, _ := r.Read(buf); uint64(n) != length {
		return nil, errors.New("Illegal Length")
	}
	if err := miscellaneous.Decode(buf, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
