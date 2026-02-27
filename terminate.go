package pgproto3

import (
	"encoding/json"
)

/*
Terminate（F）：X0004

	Byte1('X')
		标识为 termination

	Int32(4)
		消息长度，包括自己，实际内容为0，加上原始的4。
*/
type Terminate struct{}

// Frontend identifies this message as sendable by a PostgreSQL frontend.
func (*Terminate) Frontend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.

// Decode 空内容体
func (dst *Terminate) Decode(src []byte) error {
	if len(src) != 0 {
		return &invalidMessageLenErr{messageType: "Terminate", expectedLen: 0, actualLen: len(src)}
	}

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.

// Encoder 长度为4
func (src *Terminate) Encode(dst []byte) ([]byte, error) {
	return append(dst, 'X', 0, 0, 0, 4), nil
}

// MarshalJSON implements encoding/json.Marshaler.
func (src Terminate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "Terminate",
	})
}
