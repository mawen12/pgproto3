package pgproto3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/jackc/pgio"
)

/*
NotificationResponse (B)

	Byte1('A')
		标识消息是一个 notification 响应。

	Int32
		消息的字节长度，包含本身。

	Int32
		通知后端进程的进程ID。

	String
		发出通知的 channel 名称。

	String
		从通知进程传递的 "payload" 字符串。
*/
type NotificationResponse struct {
	PID     uint32
	Channel string
	Payload string
}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*NotificationResponse) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *NotificationResponse) Decode(src []byte) error {
	buf := bytes.NewBuffer(src)

	pid := binary.BigEndian.Uint32(buf.Next(4))

	b, err := buf.ReadBytes(0)
	if err != nil {
		return err
	}
	channel := string(b[:len(b)-1])

	b, err = buf.ReadBytes(0)
	if err != nil {
		return err
	}
	payload := string(b[:len(b)-1])

	*dst = NotificationResponse{PID: pid, Channel: channel, Payload: payload}
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *NotificationResponse) Encode(dst []byte) ([]byte, error) {
	dst, sp := beginMessage(dst, 'A')
	dst = pgio.AppendUint32(dst, src.PID)
	dst = append(dst, src.Channel...)
	dst = append(dst, 0)
	dst = append(dst, src.Payload...)
	dst = append(dst, 0)
	return finishMessage(dst, sp)
}

// MarshalJSON implements encoding/json.Marshaler.
func (src NotificationResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string
		PID     uint32
		Channel string
		Payload string
	}{
		Type:    "NotificationResponse",
		PID:     src.PID,
		Channel: src.Channel,
		Payload: src.Payload,
	})
}
