package protocol

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/infinivision/common/miscellaneous"
)

func init() {
	gob.Register([][]byte{})

	gob.Register(MessageError{})
	gob.Register(MessageSlice{})
	gob.Register(MessageArray{})
	gob.Register(MessageInteger{})
}

func NewMessageWriter(w *bufio.Writer) *messageWriter {
	return &messageWriter{w}
}

func NewMessage(name string, x interface{}) *Message {
	switch v := x.(type) {
	case error:
		return &Message{name, MessageError{v.Error()}}
	case int64:
		return &Message{name, MessageInteger{v}}
	case string:
		return &Message{name, MessageSlice{v}}
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
	case int64:
		msg = &Message{name, MessageInteger{v}}
	case string:
		msg = &Message{name, MessageSlice{v}}
	case []string:
		msg = &Message{name, MessageArray{v}}
	default:
		return TYPEERROR
	}
	data, err := Encode(msg)
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
	data, err := Encode(msg)
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
	if err := Decode(buf, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

// delroom name
func DelRoom(name string) []string {
	return []string{"delRoom", name}
}

// addroom name address
func AddRoom(name, address string) []string {
	return []string{"addRoom", name, address}
}

// rent name user
func Rent(name, user string) []string {
	return []string{"rent", name, user}
}

// rec name user
func Rec(name, user string) []string {
	return []string{"rec", name, user}
}
